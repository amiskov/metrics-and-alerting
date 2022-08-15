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
	Get(metricType string, metricName string) (models.Metrics, error)
	GetAll() []models.Metrics
	Ping(context.Context) error
	BatchUpsert([]models.Metrics) error // TODO: add ctx
	Upsert(context.Context, models.Metrics) error
}

type Repo struct {
	mx         *sync.Mutex
	Ctx        context.Context
	hashingKey []byte
	DB         Storage
}

func New(ctx context.Context, cfg *config.Config, s Storage) *Repo {
	repo := &Repo{
		mx:         new(sync.Mutex),
		Ctx:        ctx,
		hashingKey: []byte(cfg.HashingKey),
		DB:         s,
	}
	return repo
}

func (r Repo) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return r.DB.Ping(ctx)
}

func (r Repo) Get(metricType string, metricName string) (models.Metrics, error) {
	return r.DB.Get(metricType, metricName)
}

// Get all metrics from inmemory storage
func (r Repo) GetAll() []models.Metrics {
	return r.DB.GetAll()
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
	existingMetric, getErr := r.DB.Get(m.MType, m.ID)
	if getErr == nil && m.MType == models.MCounter && m.Delta != nil {
		currentDelta := *existingMetric.Delta
		*m.Delta += currentDelta

		if shouldHandleHash {
			h, err := m.GetHash(r.hashingKey)
			m.Hash = h
			if err != nil {
				log.Println("failed updating hash", err)
				return err
			}
		}
	}

	// add/replace the metric
	err := r.DB.Upsert(r.Ctx, m)
	if err != nil {
		return err
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
	err := r.DB.BatchUpsert(metrics)
	if err != nil {
		return err
	}

	return nil
}
