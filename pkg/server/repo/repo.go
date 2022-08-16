package repo

import (
	"context"
	"crypto/hmac"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/amiskov/metrics-and-alerting/pkg/logger"
	"github.com/amiskov/metrics-and-alerting/pkg/models"
)

type Storage interface {
	Ping(context.Context) error
	Get(metricType string, metricName string) (models.Metrics, error)
	GetAll() ([]models.Metrics, error)
	Update(models.Metrics) error
	BulkUpdate([]models.Metrics) error
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
		return m, fmt.Errorf("can't get metric with type `%s` and name `%s`: %w", m.MType, m.ID, err)
	}

	if len(r.hashingKey) > 0 {
		if err := r.updateHash(&m); err != nil {
			return m, fmt.Errorf("can't update hash for `%+v`: %w", m, err)
		}
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
		logger.Log(r.ctx).Error("repo: metric is invalid %v", err)
		return err
	}

	err := r.db.Update(m)
	if err != nil {
		logger.Log(r.ctx).Error("repo: update failed %v", err)
		return err
	}

	return nil
}

func (r *Repo) BulkUpdate(metrics []models.Metrics) error {
	// Reject operation if invalid metric found.
	// Probably it's better just skip invalid?
	for _, m := range metrics {
		if err := r.validate(m); err != nil {
			return err
		}
	}

	if err := r.db.BulkUpdate(metrics); err != nil {
		return fmt.Errorf("repo: bulk update failed: %w", err)
	}

	return nil
}

// ============ Not exported

func (r *Repo) updateHash(m *models.Metrics) error {
	if len(r.hashingKey) == 0 {
		return fmt.Errorf("no hashing key found")
	}
	hash, err := m.GetHash(r.hashingKey)
	if err != nil {
		return fmt.Errorf("can't actualize hash: %w", err)
	}
	m.Hash = hash
	return nil
}

func (r *Repo) checkHash(m models.Metrics) error {
	if len(r.hashingKey) == 0 || m.Hash == "" {
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
			"`%s:%s:%d`\n"+
			"A: %s\n"+
			"S: %s",
			r.hashingKey, m.ID, m.MType, *m.Delta, m.Hash, serverHash)
	}

	return nil
}

func (r *Repo) validate(incomingMetric models.Metrics) error {
	// Check type
	if incomingMetric.MType != models.MCounter && incomingMetric.MType != models.MGauge {
		return models.ErrorUnknownMetricType
	}

	// Check hash
	if len(r.hashingKey) != 0 && incomingMetric.Hash != "" {
		return r.checkHash(incomingMetric)
	}

	return nil
}
