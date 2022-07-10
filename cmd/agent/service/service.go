package service

import (
	"math/rand"
	"runtime"
	"sync"
	"time"

	"github.com/amiskov/metrics-and-alerting/internal/models"
)

type service struct {
	mx       *sync.RWMutex
	memStats *runtime.MemStats

	runtimeMetrics map[string]models.Gauge
	pollCount      models.Counter
	randomValue    models.Gauge
}

func (s *service) Run(pollInterval time.Duration) {
	ticker := time.NewTicker(pollInterval)
	for range ticker.C {
		s.UpdateAll()
	}
}

func New() *service {
	return &service{
		mx:             new(sync.RWMutex),
		memStats:       new(runtime.MemStats),
		runtimeMetrics: make(map[string]models.Gauge),
	}
}

func (m *service) UpdateAll() {
	runtime.ReadMemStats(m.memStats)

	m.mx.RLock()
	defer m.mx.RUnlock()

	m.runtimeMetrics["Alloc"] = models.Gauge(m.memStats.Alloc)
	m.runtimeMetrics["BuckHashSys"] = models.Gauge(m.memStats.BuckHashSys)
	m.runtimeMetrics["Frees"] = models.Gauge(m.memStats.Frees)
	m.runtimeMetrics["GCCPUFraction"] = models.Gauge(m.memStats.GCCPUFraction)
	m.runtimeMetrics["GCSys"] = models.Gauge(m.memStats.GCSys)
	m.runtimeMetrics["HeapAlloc"] = models.Gauge(m.memStats.HeapAlloc)
	m.runtimeMetrics["HeapIdle"] = models.Gauge(m.memStats.HeapIdle)
	m.runtimeMetrics["HeapInuse"] = models.Gauge(m.memStats.HeapInuse)
	m.runtimeMetrics["HeapObjects"] = models.Gauge(m.memStats.HeapObjects)
	m.runtimeMetrics["HeapReleased"] = models.Gauge(m.memStats.HeapReleased)
	m.runtimeMetrics["HeapSys"] = models.Gauge(m.memStats.HeapSys)
	m.runtimeMetrics["LastGC"] = models.Gauge(m.memStats.LastGC)
	m.runtimeMetrics["Lookups"] = models.Gauge(m.memStats.Lookups)
	m.runtimeMetrics["MCacheInuse"] = models.Gauge(m.memStats.MCacheInuse)
	m.runtimeMetrics["MCacheSys"] = models.Gauge(m.memStats.MCacheSys)
	m.runtimeMetrics["MSpanInuse"] = models.Gauge(m.memStats.MSpanInuse)
	m.runtimeMetrics["MSpanSys"] = models.Gauge(m.memStats.MSpanSys)
	m.runtimeMetrics["Mallocs"] = models.Gauge(m.memStats.Mallocs)
	m.runtimeMetrics["NextGC"] = models.Gauge(m.memStats.NextGC)
	m.runtimeMetrics["NumForcedGC"] = models.Gauge(m.memStats.NumForcedGC)
	m.runtimeMetrics["NumGC"] = models.Gauge(m.memStats.NumGC)
	m.runtimeMetrics["OtherSys"] = models.Gauge(m.memStats.OtherSys)
	m.runtimeMetrics["PauseTotalNs"] = models.Gauge(m.memStats.PauseTotalNs)
	m.runtimeMetrics["StackInuse"] = models.Gauge(m.memStats.StackInuse)
	m.runtimeMetrics["StackSys"] = models.Gauge(m.memStats.StackSys)
	m.runtimeMetrics["Sys"] = models.Gauge(m.memStats.Sys)
	m.runtimeMetrics["TotalAlloc"] = models.Gauge(m.memStats.TotalAlloc)

	m.pollCount++
	m.randomValue = models.Gauge(rand.Float64())
}

func (m *service) GetAllMetrics() []models.MetricRaw {
	m.mx.RLock()
	defer m.mx.RUnlock()

	var res []models.MetricRaw

	// Get Runtime Metrics
	for name, val := range m.runtimeMetrics {
		m := models.MetricRaw{
			Type:  "gauge",
			Name:  name,
			Value: val.String(),
		}
		res = append(res, m)
	}

	res = append(res, models.MetricRaw{
		Type:  "counter",
		Name:  "PollCount",
		Value: m.pollCount.String(),
	})

	res = append(res, models.MetricRaw{
		Type:  "gauge",
		Name:  "RandomValue",
		Value: m.randomValue.String(),
	})

	return res
}
