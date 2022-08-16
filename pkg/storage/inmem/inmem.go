package inmem

import (
	"context"
	"errors"
	"sort"
	"sync"

	"github.com/amiskov/metrics-and-alerting/pkg/models"
)

type DB struct {
	ctx        context.Context
	mx         *sync.Mutex
	data       map[string]models.Metrics // string is `type+name`
	hashingKey []byte
}

func New(ctx context.Context, key []byte) *DB {
	return &DB{
		ctx:        ctx,
		mx:         new(sync.Mutex),
		data:       make(map[string]models.Metrics),
		hashingKey: key,
	}
}

func (mdb DB) Ping(ctx context.Context) error {
	return nil
}

func (mdb DB) Get(metricType string, metricName string) (models.Metrics, error) {
	mdb.mx.Lock()
	metric, ok := mdb.data[metricType+metricName]
	mdb.mx.Unlock()

	if !ok {
		return metric, models.ErrorMetricNotFound
	}

	return metric, nil
}

// Get all metrics from inmemory storage
func (mdb DB) GetAll() ([]models.Metrics, error) {
	mdb.mx.Lock()
	defer mdb.mx.Unlock()

	metrics := []models.Metrics{}
	for _, m := range mdb.data {
		metrics = append(metrics, m)
	}

	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].ID < metrics[j].ID
	})

	return metrics, nil
}

func (mdb *DB) BulkUpdate(metrics []models.Metrics) error {
	// No mutex here, `.Update` is concurrently safe.
	for _, m := range metrics {
		updErr := mdb.Update(m)
		if updErr != nil {
			return updErr
		}
	}
	return nil
}

func (mdb *DB) Update(m models.Metrics) error {
	if m.MType == models.MCounter {
		existingMetric, err := mdb.Get(m.MType, m.ID)
		if err == nil {
			if existingMetric.Delta == nil || m.Delta == nil {
				return errors.New("empty Delta for counter metric")
			}

			currentDelta := *existingMetric.Delta
			newDelta := currentDelta + *m.Delta
			m.Delta = &newDelta
		}
	}

	mdb.mx.Lock()
	mdb.data[m.MType+m.ID] = m
	mdb.mx.Unlock()

	return nil
}
