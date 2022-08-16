package repo

import (
	"context"
	"crypto/hmac"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/amiskov/metrics-and-alerting/pkg/models"
)

type Storage interface {
	Ping(context.Context) error
	Get(metricType string, metricName string) (models.Metrics, error)
	GetAll() ([]models.Metrics, error)
	Upsert(context.Context, models.Metrics) error
	BatchUpsert([]models.Metrics) error // TODO: add ctx
}

type Repo struct {
	mx         *sync.Mutex
	ctx        context.Context
	hashingKey []byte
	db         Storage
}

func New(ctx context.Context, hashingKey []byte, s Storage) *Repo {
	repo := &Repo{
		mx:         new(sync.Mutex),
		ctx:        ctx,
		hashingKey: hashingKey,
		db:         s,
	}
	return repo
}

func (r Repo) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return r.db.Ping(ctx)
}

func (r Repo) Get(metricType string, metricName string) (models.Metrics, error) {
	m, err := r.db.Get(metricType, metricName)
	if err != nil {
		return m, err
	}
	if err := r.updateHash(&m); err != nil {
		return m, err
	}
	return m, nil
}

// Get all metrics from inmemory storage
func (r Repo) GetAll() ([]models.Metrics, error) {
	metrics, err := r.db.GetAll()
	if err != nil {
		return metrics, err
	}

	if len(r.hashingKey) > 0 {
		for k := range metrics {
			if err := r.updateHash(&metrics[k]); err != nil {
				return metrics, err
			}
		}
	}

	return metrics, nil
}

func (r *Repo) Update(m models.Metrics) error {
	if err := r.validate(m); err != nil {
		return err
	}

	if m.MType == models.MCounter {
		err := r.updateDelta(&m)
		if err != nil {
			return err
		}
	}

	// add/replace the metric
	err := r.db.Upsert(r.ctx, m)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repo) BatchUpsert(metrics []models.Metrics) error {
	for _, m := range metrics {
		r.Update(m)
	}
	return nil
}

// ============ Not exported

// Updates the Delta of a `counter` metric if metric already exists.
func (r *Repo) updateDelta(m *models.Metrics) error {
	existingMetric, err := r.db.Get(m.MType, m.ID)
	if err != nil { // Metric not exists, nothing to update
		return nil
	}

	if existingMetric.Delta == nil || m.Delta == nil {
		return errors.New("empty Delta for counter metric")
	}

	currentDelta := *existingMetric.Delta
	newDelta := currentDelta + *m.Delta
	m.Delta = &newDelta
	return nil
}

func (r *Repo) updateHash(m *models.Metrics) error {
	if len(r.hashingKey) == 0 {
		return nil
	}
	hash, err := m.GetHash(r.hashingKey)
	if err != nil {
		return fmt.Errorf("can't actualize hash: %w", err)
	}
	m.Hash = hash
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

func (r *Repo) validate(incomingMetric models.Metrics) error {
	// Check type
	if incomingMetric.MType != models.MCounter && incomingMetric.MType != models.MGauge {
		return models.ErrorUnknownMetricType
	}

	// Check hash
	if len(r.hashingKey) != 0 {
		return r.checkHash(incomingMetric)
	}

	return nil
}
