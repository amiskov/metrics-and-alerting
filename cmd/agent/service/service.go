package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"runtime"
	"sync"
	"time"

	"github.com/amiskov/metrics-and-alerting/cmd/server/repo/inmem"
	"github.com/amiskov/metrics-and-alerting/internal/models"
)

type service struct {
	mx         *sync.RWMutex
	memStats   *runtime.MemStats
	metrics    *inmem.DB
	hashingKey []byte
}

func New(key []byte) *service {
	return &service{
		mx:         new(sync.RWMutex),
		memStats:   new(runtime.MemStats),
		metrics:    inmem.NewInmemDB(),
		hashingKey: key,
	}
}

func (s *service) Run(ctx context.Context, done chan bool, pollInterval time.Duration) {
	ticker := time.NewTicker(pollInterval)

	go func() {
		<-ctx.Done()
		ticker.Stop()
		log.Println("Metrics update stopped.")
		done <- true
	}()

	for range ticker.C {
		s.updateMetrics()
	}
}

func (s *service) GetMetrics() []models.Metrics {
	s.mx.Lock()
	defer s.mx.Unlock()
	return s.metrics.GetAll()
}

func (s *service) updateMetrics() {
	runtime.ReadMemStats(s.memStats)

	s.mx.Lock()
	defer s.mx.Unlock()

	s.updateGauge("Alloc", float64(s.memStats.Alloc))
	s.updateGauge("BuckHashSys", float64(s.memStats.BuckHashSys))
	s.updateGauge("Frees", float64(s.memStats.Frees))
	s.updateGauge("GCCPUFraction", s.memStats.GCCPUFraction)
	s.updateGauge("GCSys", float64(s.memStats.GCSys))
	s.updateGauge("HeapAlloc", float64(s.memStats.HeapAlloc))
	s.updateGauge("HeapIdle", float64(s.memStats.HeapIdle))
	s.updateGauge("HeapInuse", float64(s.memStats.HeapInuse))
	s.updateGauge("HeapObjects", float64(s.memStats.HeapObjects))
	s.updateGauge("HeapReleased", float64(s.memStats.HeapReleased))
	s.updateGauge("HeapSys", float64(s.memStats.HeapSys))
	s.updateGauge("LastGC", float64(s.memStats.LastGC))
	s.updateGauge("Lookups", float64(s.memStats.Lookups))
	s.updateGauge("MCacheInuse", float64(s.memStats.MCacheInuse))
	s.updateGauge("MCacheSys", float64(s.memStats.MCacheSys))
	s.updateGauge("MSpanInuse", float64(s.memStats.MSpanInuse))
	s.updateGauge("MSpanSys", float64(s.memStats.MSpanSys))
	s.updateGauge("Mallocs", float64(s.memStats.Mallocs))
	s.updateGauge("NextGC", float64(s.memStats.NextGC))
	s.updateGauge("NumForcedGC", float64(s.memStats.NumForcedGC))
	s.updateGauge("NumGC", float64(s.memStats.NumGC))
	s.updateGauge("OtherSys", float64(s.memStats.OtherSys))
	s.updateGauge("PauseTotalNs", float64(s.memStats.PauseTotalNs))
	s.updateGauge("StackInuse", float64(s.memStats.StackInuse))
	s.updateGauge("StackSys", float64(s.memStats.StackSys))
	s.updateGauge("Sys", float64(s.memStats.Sys))
	s.updateGauge("TotalAlloc", float64(s.memStats.TotalAlloc))

	s.updateCounter("PollCount")

	s.updateGauge("RandomValue", rand.Float64()) // nolint: gosec

	log.Println("Metrics has been updated.")
}

func (s *service) updateCounter(id string) {
	m, err := s.metrics.Get(models.MCounter, id)
	if errors.Is(err, models.ErrorMetricNotFound) {
		var zero int64
		m = models.Metrics{
			ID:    id,
			MType: models.MCounter,
			Delta: &zero,
		}
		fmt.Printf("First update! %v, %d\n", m, *m.Delta)
	}

	log.Printf("m.before +1: %s:%d, %s.\n", m.ID, *m.Delta, m.Hash)

	// if metric exists, increment its Delta
	delta := *m.Delta + 1
	m.Delta = &delta

	log.Printf("m.before hash: %s:%d.\n", m.ID, *m.Delta)

	// refresh metric hash if key available
	if len(s.hashingKey) != 0 {
		hash, err := m.GetHash(s.hashingKey)
		if err != nil {
			log.Printf("failed creating hash for %s: %v", id, err)
		}
		m.Hash = hash
	}

	log.Printf("m.after hash: %s:%d, %s.\n", m.ID, *m.Delta, m.Hash)

	// add/replace metric in storage
	updErr := s.metrics.Upsert(m)
	if updErr != nil {
		log.Printf("can't update %+v\n", m)
	}
}

func (s *service) updateGauge(id string, val float64) {
	m := models.Metrics{
		ID:    id,
		MType: models.MGauge,
		Value: &val,
	}

	var hashingErr error
	m.Hash, hashingErr = m.GetHash(s.hashingKey)
	if hashingErr != nil {
		log.Printf("failed creating hash for %s: %v", id, hashingErr)
	}

	updErr := s.metrics.Upsert(m)
	if updErr != nil {
		log.Printf("can't update %+v\n", m)
	}
}
