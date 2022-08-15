package repo

import (
	"context"
	"crypto/hmac"
	"encoding/hex"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/amiskov/metrics-and-alerting/cmd/server/config"
	"github.com/amiskov/metrics-and-alerting/cmd/server/repo/inmem"
	"github.com/amiskov/metrics-and-alerting/internal/models"
)

type Storage interface {
	Ping(context.Context) error
	Dump(context.Context, []models.Metrics) error
	Restore() ([]models.Metrics, error)
	BatchUpsert([]models.Metrics) error // TODO: add ctx
}

// Repo keeps metrics inmemory, dumps metrics to persistent `Storage` intervally,
// and preloads metrics from `Storage` if `Restore` is `true`.
type Repo struct {
	mx            *sync.Mutex
	inmemDB       *inmem.DB
	StoreInterval time.Duration // store immediately if `0`
	Restore       bool          // restore from persistent storage on start if `true`
	Ctx           context.Context
	Finished      chan bool // to make sure we wrote the data while terminating
	hashingKey    []byte
	DB            Storage
}

func New(ctx context.Context, finished chan bool, cfg *config.Config, s Storage) *Repo {
	repo := &Repo{
		mx:            new(sync.Mutex),
		inmemDB:       inmem.NewInmemDB(),
		StoreInterval: cfg.StoreInterval,
		Restore:       cfg.Restore,
		Ctx:           ctx,
		Finished:      finished,
		hashingKey:    []byte(cfg.HashingKey),
		DB:            s,
	}

	// TODO: fixme. This is a cheat for inc11 to skip interval dumping for Postgres.
	// When using Pg we persist data on every update.
	if cfg.PgDSN != "" {
		repo.StoreInterval = 0
	}

	// Restore metrics from persistent storage or create an empty inmemory DB
	if cfg.Restore {
		restoredMetrics, err := repo.DB.Restore()
		if err != nil {
			log.Println("can't restore data from storage:", err)
		}

		errR := repo.inmemDB.BatchUpsert(restoredMetrics)
		if errR != nil {
			log.Println("cant restore metrics to inmem")
		}

		// create hashes for running server with current hashing key
		if len(repo.hashingKey) != 0 {
			// NB: Since tests run each time with a different key we can't store hashes in persistent DB.
			err := repo.inmemDB.ActualizeHashes(repo.hashingKey)
			if err != nil {
				log.Println("can't update hashes", err)
			}
		}
	}

	var ticker *time.Ticker

	go repo.HandleTermination(ticker)

	if repo.StoreInterval > 0 {
		ticker = time.NewTicker(repo.StoreInterval)
		go repo.SavePeriodically(ticker)
	}

	return repo
}

// Handle program termination: save data and stop ticker if it's running.
func (r Repo) HandleTermination(ticker *time.Ticker) {
	<-r.Ctx.Done()
	if ticker != nil {
		ticker.Stop()
		log.Println("Saving timer stopped.")
	}
	if err := r.DB.Dump(r.Ctx, r.inmemDB.GetAll()); err != nil {
		log.Println("failed saving to persistent storage:", err)
	}
	r.Finished <- true
}

// Interval saving to persistent storage.
func (r Repo) SavePeriodically(ticker *time.Ticker) {
	if r.StoreInterval <= 0 {
		log.Println("interval should be non-zero positive number")
		return
	}
	defer ticker.Stop()
	for range ticker.C {
		if err := r.DB.Dump(r.Ctx, r.inmemDB.GetAll()); err != nil {
			log.Println("interval saving saving to persistent storage failed:", err)
		}
	}
}

func (r Repo) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return r.DB.Ping(ctx)
}

func (r Repo) Get(metricType string, metricName string) (models.Metrics, error) {
	return r.inmemDB.Get(metricType, metricName)
}

// Get all metrics from inmemory storage
func (r Repo) GetAll() []models.Metrics {
	return r.inmemDB.GetAll()
}

func (r *Repo) Update(m models.Metrics) error {
	// Check metric type
	if m.MType != models.MCounter && m.MType != models.MGauge {
		return models.ErrorUnknownMetricType
	}

	shouldHandleHash := m.Hash != "" && len(r.hashingKey) != 0

	// Check metric hash
	if shouldHandleHash {
		if err := r.checkHash(m); err != nil {
			v := ""

			if m.MType == models.MCounter {
				v = fmt.Sprintf("%d", *m.Delta)
			}
			if m.MType == models.MGauge {
				v = fmt.Sprintf("%f", *m.Value)
			}

			log.Printf("bad hash `%s` for `%s:%s:%s` (`%s`). Error: %v.\n",
				m.Hash, m.ID, m.MType, v, string(r.hashingKey), err)
			return err
		}
	}

	// For `counter` metrics, update the Delta if metric already exists
	existingMetric, getErr := r.inmemDB.Get(m.MType, m.ID)
	if getErr == nil && m.MType == models.MCounter {
		currentDelta := *existingMetric.Delta
		*m.Delta += currentDelta

		if shouldHandleHash {
			var err error
			m.Hash, err = m.GetHash(r.hashingKey)
			if err != nil {
				log.Println("failed updating hash", err)
				return err
			}
		}
	}

	// add/replace the metric
	err := r.inmemDB.Upsert(m)
	if err != nil {
		return err
	}

	// Immediately save metrics to the persistent storage if needed
	if r.StoreInterval == 0 {
		// TODO: use Upsert here
		if err := r.DB.Dump(r.Ctx, r.inmemDB.GetAll()); err != nil {
			log.Println("failed saving metrics to persistent storage.", err)
		}
	}

	return nil
}

func (r *Repo) checkHash(m models.Metrics) error {
	if m.Hash == "" {
		return nil // nothing to check
	}

	metricHash, err := hex.DecodeString(m.Hash)
	if err != nil {
		return fmt.Errorf(`bad agent hash: %w`, err)
	}

	serverHash, err := m.GetHash(r.hashingKey)
	if err != nil {
		return fmt.Errorf("failed creating server hash: %w", err)
	}

	seHex, err := hex.DecodeString(serverHash)
	if err != nil {
		return fmt.Errorf("bad server hash: %w", err)
	}

	if !hmac.Equal(metricHash, seHex) {
		return fmt.Errorf("agent and server hashes are not equal.\n"+
			"Server key: %s\n"+
			"A: %s\n"+
			"S: %s",
			r.hashingKey, m.Hash, serverHash)
	}

	return nil
}

func (r *Repo) BatchUpsert(metrics []models.Metrics) error {
	// TODO: check types and hashes. See `Update`, create separate functions for metrics validation.
	// TODO: For `counter` metrics, update the Delta if metric already exists

	// add/replace metrics
	err := r.inmemDB.BatchUpsert(metrics)
	if err != nil {
		return err
	}

	// Immediately save metrics to the persistent storage if needed
	if r.StoreInterval == 0 {
		if err := r.DB.Dump(r.Ctx, metrics); err != nil {
			log.Println("failed saving metrics to persistent storage.", err)
		}
	}

	return nil
}
