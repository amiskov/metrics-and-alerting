package inmem

import (
	"log"
	"sort"
	"sync"

	"github.com/amiskov/metrics-and-alerting/internal/models"
)

type DB struct {
	mx   *sync.Mutex
	data map[string]models.Metrics // string is `type+name`
}

func NewInmemDB() *DB {
	return &DB{
		mx:   new(sync.Mutex),
		data: make(map[string]models.Metrics),
	}
}

func (mdb DB) Get(metricType string, metricName string) (models.Metrics, error) {
	// Handle wrong metric type
	if metricType != models.MCounter && metricType != models.MGauge {
		return models.Metrics{}, models.ErrorMetricNotFound
	}

	mdb.mx.Lock()
	metric, ok := mdb.data[metricType+metricName]
	mdb.mx.Unlock()

	if !ok {
		return metric, models.ErrorMetricNotFound
	}

	return metric, nil
}

// Get all metrics from inmemory storage
func (mdb DB) GetAll() []models.Metrics {
	mdb.mx.Lock()
	defer mdb.mx.Unlock()

	metrics := []models.Metrics{}
	for _, m := range mdb.data {
		metrics = append(metrics, m)
	}

	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].ID < metrics[j].ID
	})

	return metrics
}

func (mdb *DB) BatchUpsert(metrics []models.Metrics) error {
	for _, m := range metrics {
		// `.Upsert` is concurrently safe
		updErr := mdb.Upsert(m)
		if updErr != nil {
			return updErr
		}
	}

	return nil
}

func (mdb *DB) Upsert(m models.Metrics) error {
	mdb.mx.Lock()
	defer mdb.mx.Unlock()

	mdb.data[m.MType+m.ID] = m

	return nil
}

func (mdb *DB) ActualizeHashes(key []byte) error {
	mdb.mx.Lock()
	defer mdb.mx.Unlock()

	for k, m := range mdb.data {
		hash, err := m.GetHash(key)
		if err != nil {
			log.Println("can't actualize hash", err)
			return err
		}
		m.Hash = hash
		mdb.data[k] = m
	}
	return nil
}
