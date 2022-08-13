package repo

import (
	"context"
	"crypto/hmac"
	"encoding/hex"
	"log"
	"time"

	"github.com/amiskov/metrics-and-alerting/cmd/server/config"
	"github.com/amiskov/metrics-and-alerting/internal/models"
)

type Storage interface {
	Ping(context.Context) error
	Get(metricType string, metricName string) (models.Metrics, error)
	GetAll() []models.Metrics
	Update(m models.Metrics) error
	Dump() error
}

type Repo struct {
	inmemDB       models.MetricsDB
	StoreInterval time.Duration // store immediately if `0`
	Restore       bool          // restore from file on start if `true`
	Ctx           context.Context
	Finished      chan bool // to make sure we wrote the data while terminating
	HashingKey    []byte
	DB            Storage
}

func New(ctx context.Context, finished chan bool, cfg *config.Config, s Storage) *Repo {
	repo := &Repo{
		// Handle inmemory storage inside repo!
		inmemDB:       make(models.MetricsDB),
		StoreInterval: cfg.StoreInterval,
		Restore:       cfg.Restore,
		Ctx:           ctx,
		Finished:      finished,
		HashingKey:    []byte(cfg.HashingKey),
		DB:            s,
	}

	repo.SavePeriodically()

	return repo
}

func (r Repo) SavePeriodically() {
	var ticker *time.Ticker
	save := func() {
		if err := r.DB.Dump(); err != nil {
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

	return r.DB.Get(metricType, metricName)
}

func (r Repo) GetAll() []models.Metrics {
	return r.DB.GetAll()
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

	return r.DB.Update(m)
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
