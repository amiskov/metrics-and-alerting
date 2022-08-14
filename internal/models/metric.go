package models

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"log"
	"sort"
	"sync"
)

const (
	MGauge   = "gauge"
	MCounter = "counter"
)

type (
	Gauge   float64
	Counter int64
)

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
	Hash  string   `json:"hash,omitempty"`  // значение хеш-функции
}

func (m Metrics) GetHash(key []byte) (string, error) {
	var src string

	switch m.MType {
	case MCounter:
		src = fmt.Sprintf("%s:%s:%d", m.ID, m.MType, *m.Delta)
	case MGauge:
		src = fmt.Sprintf("%s:%s:%f", m.ID, m.MType, *m.Value)
	default:
		return src, ErrorUnknownMetricType
	}

	h := hmac.New(sha256.New, key)
	h.Write([]byte(src))

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

type InmemDB struct {
	mx   *sync.Mutex
	data map[string]Metrics // string is `type+name`
}

func NewInmemDB() *InmemDB {
	return &InmemDB{
		mx:   new(sync.Mutex),
		data: make(map[string]Metrics),
	}
}

func (mdb InmemDB) Get(metricType string, metricName string) (Metrics, error) {
	// Handle wrong metric type
	if metricType != MCounter && metricType != MGauge {
		return Metrics{}, ErrorMetricNotFound
	}

	mdb.mx.Lock()
	metric, ok := mdb.data[metricType+metricName]
	mdb.mx.Unlock()

	if !ok {
		return metric, ErrorMetricNotFound
	}

	return metric, nil
}

// Get all metrics from inmemory storage
func (mdb InmemDB) GetAll() []Metrics {
	mdb.mx.Lock()
	defer mdb.mx.Unlock()

	metrics := []Metrics{}
	for _, m := range mdb.data {
		metrics = append(metrics, m)
	}

	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].ID < metrics[j].ID
	})

	return metrics
}

func (mdb *InmemDB) Upsert(m Metrics) error {
	mdb.mx.Lock()
	defer mdb.mx.Unlock()

	mdb.data[m.MType+m.ID] = m

	return nil
}

func (mdb *InmemDB) ActualizeHashes(key []byte) error {
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
