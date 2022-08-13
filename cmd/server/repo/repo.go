package repo

import (
	"context"
	"crypto/hmac"
	"encoding/hex"
	"log"
	"sort"
	"sync"
	"time"

	"github.com/amiskov/metrics-and-alerting/cmd/server/config"
	"github.com/amiskov/metrics-and-alerting/internal/models"
)

type Storage interface {
	Ping(context.Context) error
	Dump(models.MetricsDB) error
	Restore(models.MetricsDB) error
}

// Repo keeps metrics inmemory, dumps metrics to persistent `Storage` intervally,
// and preloads metrics from `Storage` if `Restore` is `true`.
type Repo struct {
	mx *sync.Mutex

	// TODO: Update, Get and other methods could be handled as methods of the MetricsDB itself.
	inmemDB models.MetricsDB

	StoreInterval time.Duration // store immediately if `0`
	Restore       bool          // restore from persistent storage on start if `true`
	Ctx           context.Context
	Finished      chan bool // to make sure we wrote the data while terminating
	HashingKey    []byte
	DB            Storage
}

func New(ctx context.Context, finished chan bool, cfg *config.Config, s Storage) *Repo {
	repo := &Repo{
		mx:            new(sync.Mutex),
		inmemDB:       make(models.MetricsDB),
		StoreInterval: cfg.StoreInterval,
		Restore:       cfg.Restore,
		Ctx:           ctx,
		Finished:      finished,
		HashingKey:    []byte(cfg.HashingKey),
		DB:            s,
	}

	// Restore from file or create empty metrics DB
	if cfg.Restore {
		err := repo.DB.Restore(repo.inmemDB)
		if err != nil {
			log.Println("can't restore data from storage:", err)
		}
	}

	repo.SavePeriodically()

	return repo
}

func (r Repo) SavePeriodically() {
	var ticker *time.Ticker
	save := func() {
		if err := r.DB.Dump(r.inmemDB); err != nil {
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
	// Handle wrong metric type
	if metricType != models.MCounter && metricType != models.MGauge {
		return models.Metrics{}, models.ErrorMetricNotFound
	}

	r.mx.Lock()
	metric, ok := r.inmemDB[metricName]
	r.mx.Unlock()

	if !ok {
		return metric, models.ErrorMetricNotFound
	}

	return metric, nil
}

// Get all metrics from inmemory storage
func (r Repo) GetAll() []models.Metrics {
	r.mx.Lock()
	defer r.mx.Unlock()

	metrics := []models.Metrics{}
	for _, m := range r.inmemDB {
		metrics = append(metrics, m)
	}

	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].ID < metrics[j].ID
	})

	return metrics
}

// Updates the inmem values.
func (r *Repo) update(m models.Metrics) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	// Metric not exists on the server
	if _, ok := r.inmemDB[m.ID]; !ok {
		r.inmemDB[m.ID] = m
		return nil
	}

	updatedMetric := m

	// Update existing counter metric
	if m.MType == models.MCounter {
		delta := *r.inmemDB[m.ID].Delta + *m.Delta
		updatedMetric.Delta = &delta

		if m.Hash != "" {
			newHash, err := updatedMetric.GetHash(r.HashingKey)
			if err != nil {
				return err
			}
			updatedMetric.Hash = newHash
		}
	}

	r.inmemDB[m.ID] = updatedMetric

	return nil
}

func (r *Repo) Update(m models.Metrics) error {
	// Check metric type
	if m.MType != models.MCounter && m.MType != models.MGauge {
		return models.ErrorUnknownMetricType
	}

	// Check metric hash
	if m.Hash != "" {
		if err := r.checkHash(m); err != nil {
			return err
		}
	}

	err := r.update(m) // inmemory update
	if err != nil {
		return err
	}

	// Immediately save metrics to the persistent storage.
	if r.StoreInterval == 0 {
		return r.DB.Dump(r.inmemDB)
	}

	return nil
}

func (r *Repo) checkHash(m models.Metrics) error {
	if m.Hash == "" {
		return nil // nothing to check
	}

	metricHash, err := hex.DecodeString(m.Hash)
	if err != nil {
		log.Println("bad agent hash", err)
		return models.ErrorBadMetricFormat
	}

	serverHash, err := m.GetHash(r.HashingKey)
	if err != nil {
		log.Println("failed creating server hash", err)
		return models.ErrorBadMetricFormat
	}

	seHex, err := hex.DecodeString(serverHash)
	if err != nil {
		log.Println("bad server hash", err)
		return models.ErrorBadMetricFormat
	}

	if !hmac.Equal(metricHash, seHex) {
		return models.ErrorBadMetricFormat
	}

	return nil
}
