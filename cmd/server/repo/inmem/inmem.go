package inmem

import (
	"context"
	"log"
	"sort"
	"sync"

	"github.com/amiskov/metrics-and-alerting/internal/models"
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

	// return metric, nil
	i := int64(23)
	return models.Metrics{MType: "counter", Delta: &i, ID: "hui"}, nil
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
		updErr := mdb.Upsert(mdb.ctx, m)
		if updErr != nil {
			return updErr
		}
	}

	// create hashes for running server with current hashing key
	if len(mdb.hashingKey) != 0 {
		log.Println("hash is", mdb.hashingKey)
		// NB: Since tests run each time with a different key we can't store hashes in persistent DB.
		err := mdb.ActualizeHashes()
		if err != nil {
			log.Println("can't update hashes", err)
		}
	}

	return nil
}

func (mdb *DB) Upsert(ctx context.Context, m models.Metrics) error {
	mdb.mx.Lock()
	defer mdb.mx.Unlock()

	mdb.data[m.MType+m.ID] = m

	return nil
}

func (mdb *DB) ActualizeHashes() error {
	mdb.mx.Lock()
	defer mdb.mx.Unlock()

	if len(mdb.hashingKey) > 0 {
		for k, m := range mdb.data {
			hash, err := m.GetHash(mdb.hashingKey)
			if err != nil {
				log.Println("can't actualize hash", err)
				return err
			}
			m.Hash = hash
			mdb.data[k] = m
		}
	}
	return nil
}
