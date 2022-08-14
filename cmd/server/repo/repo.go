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
	"github.com/amiskov/metrics-and-alerting/internal/models"
)

type Storage interface {
	Ping(context.Context) error
	Dump(models.InmemDB) error
	Restore(models.InmemDB) error
}

// Repo keeps metrics inmemory, dumps metrics to persistent `Storage` intervally,
// and preloads metrics from `Storage` if `Restore` is `true`.
type Repo struct {
	mx            *sync.Mutex
	inmemDB       *models.InmemDB
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
		inmemDB:       models.NewInmemDB(),
		StoreInterval: cfg.StoreInterval,
		Restore:       cfg.Restore,
		Ctx:           ctx,
		Finished:      finished,
		hashingKey:    []byte(cfg.HashingKey),
		DB:            s,
	}

	// Restore from file or create empty metrics DB
	if cfg.Restore {
		err := repo.DB.Restore(*repo.inmemDB)
		if err != nil {
			log.Println("can't restore data from storage:", err)
		}

		// create hashes for running server with current hashing key
		if len(repo.hashingKey) != 0 {
			err := repo.inmemDB.ActualizeHashes(repo.hashingKey)
			if err != nil {
				log.Println("can't update hashes", err)
			}
		}
	}

	repo.SavePeriodically()

	return repo
}

func (r Repo) SavePeriodically() {
	var ticker *time.Ticker
	save := func() {
		if err := r.DB.Dump(*r.inmemDB); err != nil {
			log.Println("Failed saving metrics.", err)
			return
		}
	}

	// Interval saving to file if interval is not `0`.
	if r.StoreInterval > 0 {
		go func() {
			ticker = time.NewTicker(r.StoreInterval)
			defer ticker.Stop()
			for range ticker.C {
				log.Println("saving...")
				save()
			}
		}()
	}

	// Handle terminating message: save data and stop ticker if it's running.
	go func() {
		<-r.Ctx.Done()
		if ticker != nil {
			ticker.Stop()
			log.Println("Saving timer stopped.")
		}
		save()
		log.Println("Data saved.")
		r.Finished <- true
	}()
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
		return r.DB.Dump(*r.inmemDB)
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
