package service

import (
	"context"
	"log"
	"math/rand"
	"runtime"
	"sync"
	"time"

	"github.com/amiskov/metrics-and-alerting/internal/common"
	"github.com/amiskov/metrics-and-alerting/internal/models"
)

type service struct {
	mx       *sync.RWMutex
	memStats *runtime.MemStats

	metricsDB models.MetricsDB

	runtimeMetrics map[string]models.Gauge
	pollCount      models.Counter
	randomValue    models.Gauge
	hashingKey     []byte
}

func New(key []byte) *service {
	return &service{
		mx:             new(sync.RWMutex),
		memStats:       new(runtime.MemStats),
		runtimeMetrics: make(map[string]models.Gauge),
		hashingKey:     key,
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

	res := []models.Metrics{}

	// Get Runtime Metrics
	for name, val := range s.runtimeMetrics {
		val := float64(val)
		m := models.Metrics{
			MType: models.MGauge,
			ID:    name,
			Value: &val,
		}
		hash, _ := common.Hash(m, s.hashingKey)
		m.Hash = hash
		res = append(res, m)
	}

	val := int64(s.pollCount)
	pcm := models.Metrics{
		MType: models.MCounter,
		ID:    "PollCount",
		Delta: &val,
	}
	hash, _ := common.Hash(pcm, s.hashingKey)
	pcm.Hash = hash
	res = append(res, pcm)

	randVal := float64(s.randomValue)
	rm := models.Metrics{
		MType: models.MGauge,
		ID:    "RandomValue",
		Value: &randVal,
	}
	hash, _ = common.Hash(rm, s.hashingKey)
	rm.Hash = hash
	res = append(res, rm)

	return res
}

func (s *service) updateMetrics() {
	runtime.ReadMemStats(s.memStats)

	s.mx.Lock()
	defer s.mx.Unlock()

	s.runtimeMetrics["Alloc"] = models.Gauge(s.memStats.Alloc)
	s.runtimeMetrics["BuckHashSys"] = models.Gauge(s.memStats.BuckHashSys)
	s.runtimeMetrics["Frees"] = models.Gauge(s.memStats.Frees)
	s.runtimeMetrics["GCCPUFraction"] = models.Gauge(s.memStats.GCCPUFraction)
	s.runtimeMetrics["GCSys"] = models.Gauge(s.memStats.GCSys)
	s.runtimeMetrics["HeapAlloc"] = models.Gauge(s.memStats.HeapAlloc)
	s.runtimeMetrics["HeapIdle"] = models.Gauge(s.memStats.HeapIdle)
	s.runtimeMetrics["HeapInuse"] = models.Gauge(s.memStats.HeapInuse)
	s.runtimeMetrics["HeapObjects"] = models.Gauge(s.memStats.HeapObjects)
	s.runtimeMetrics["HeapReleased"] = models.Gauge(s.memStats.HeapReleased)
	s.runtimeMetrics["HeapSys"] = models.Gauge(s.memStats.HeapSys)
	s.runtimeMetrics["LastGC"] = models.Gauge(s.memStats.LastGC)
	s.runtimeMetrics["Lookups"] = models.Gauge(s.memStats.Lookups)
	s.runtimeMetrics["MCacheInuse"] = models.Gauge(s.memStats.MCacheInuse)
	s.runtimeMetrics["MCacheSys"] = models.Gauge(s.memStats.MCacheSys)
	s.runtimeMetrics["MSpanInuse"] = models.Gauge(s.memStats.MSpanInuse)
	s.runtimeMetrics["MSpanSys"] = models.Gauge(s.memStats.MSpanSys)
	s.runtimeMetrics["Mallocs"] = models.Gauge(s.memStats.Mallocs)
	s.runtimeMetrics["NextGC"] = models.Gauge(s.memStats.NextGC)
	s.runtimeMetrics["NumForcedGC"] = models.Gauge(s.memStats.NumForcedGC)
	s.runtimeMetrics["NumGC"] = models.Gauge(s.memStats.NumGC)
	s.runtimeMetrics["OtherSys"] = models.Gauge(s.memStats.OtherSys)
	s.runtimeMetrics["PauseTotalNs"] = models.Gauge(s.memStats.PauseTotalNs)
	s.runtimeMetrics["StackInuse"] = models.Gauge(s.memStats.StackInuse)
	s.runtimeMetrics["StackSys"] = models.Gauge(s.memStats.StackSys)
	s.runtimeMetrics["Sys"] = models.Gauge(s.memStats.Sys)
	s.runtimeMetrics["TotalAlloc"] = models.Gauge(s.memStats.TotalAlloc)

	s.pollCount++
	s.randomValue = models.Gauge(rand.Float64()) // nolint: gosec
	log.Println("Metrics has been updated.")
}
