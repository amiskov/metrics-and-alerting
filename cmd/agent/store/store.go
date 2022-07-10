package store

import (
	"math/rand"
	"runtime"
	"sync"

	"github.com/amiskov/metrics-and-alerting/internal/models"
)

type metrics struct {
	mx       *sync.RWMutex
	memStats *runtime.MemStats

	RuntimeMetrics map[string]models.Gauge
	PollCount      models.Counter
	RandomValue    models.Gauge
}

func NewMetricsStore() *metrics {
	return &metrics{
		mx:             new(sync.RWMutex),
		memStats:       new(runtime.MemStats),
		RuntimeMetrics: make(map[string]models.Gauge),
	}
}

func (m *metrics) UpdateAll() {
	runtime.ReadMemStats(m.memStats)

	m.mx.RLock()
	defer m.mx.RUnlock()

	m.RuntimeMetrics["Alloc"] = models.Gauge(m.memStats.Alloc)
	m.RuntimeMetrics["BuckHashSys"] = models.Gauge(m.memStats.BuckHashSys)
	m.RuntimeMetrics["Frees"] = models.Gauge(m.memStats.Frees)
	m.RuntimeMetrics["GCCPUFraction"] = models.Gauge(m.memStats.GCCPUFraction)
	m.RuntimeMetrics["GCSys"] = models.Gauge(m.memStats.GCSys)
	m.RuntimeMetrics["HeapAlloc"] = models.Gauge(m.memStats.HeapAlloc)
	m.RuntimeMetrics["HeapIdle"] = models.Gauge(m.memStats.HeapIdle)
	m.RuntimeMetrics["HeapInuse"] = models.Gauge(m.memStats.HeapInuse)
	m.RuntimeMetrics["HeapObjects"] = models.Gauge(m.memStats.HeapObjects)
	m.RuntimeMetrics["HeapReleased"] = models.Gauge(m.memStats.HeapReleased)
	m.RuntimeMetrics["HeapSys"] = models.Gauge(m.memStats.HeapSys)
	m.RuntimeMetrics["LastGC"] = models.Gauge(m.memStats.LastGC)
	m.RuntimeMetrics["Lookups"] = models.Gauge(m.memStats.Lookups)
	m.RuntimeMetrics["MCacheInuse"] = models.Gauge(m.memStats.MCacheInuse)
	m.RuntimeMetrics["MCacheSys"] = models.Gauge(m.memStats.MCacheSys)
	m.RuntimeMetrics["MSpanInuse"] = models.Gauge(m.memStats.MSpanInuse)
	m.RuntimeMetrics["MSpanSys"] = models.Gauge(m.memStats.MSpanSys)
	m.RuntimeMetrics["Mallocs"] = models.Gauge(m.memStats.Mallocs)
	m.RuntimeMetrics["NextGC"] = models.Gauge(m.memStats.NextGC)
	m.RuntimeMetrics["NumForcedGC"] = models.Gauge(m.memStats.NumForcedGC)
	m.RuntimeMetrics["NumGC"] = models.Gauge(m.memStats.NumGC)
	m.RuntimeMetrics["OtherSys"] = models.Gauge(m.memStats.OtherSys)
	m.RuntimeMetrics["PauseTotalNs"] = models.Gauge(m.memStats.PauseTotalNs)
	m.RuntimeMetrics["StackInuse"] = models.Gauge(m.memStats.StackInuse)
	m.RuntimeMetrics["StackSys"] = models.Gauge(m.memStats.StackSys)
	m.RuntimeMetrics["Sys"] = models.Gauge(m.memStats.Sys)
	m.RuntimeMetrics["TotalAlloc"] = models.Gauge(m.memStats.TotalAlloc)

	m.PollCount++
	m.RandomValue = models.Gauge(rand.Float64())
}

func (m *metrics) GetAll() []models.MetricRaw {
	m.mx.RLock()
	defer m.mx.RUnlock()

	var res []models.MetricRaw

	// Get Runtime Metrics
	for name, val := range m.RuntimeMetrics {
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
		Value: m.PollCount.String(),
	})

	res = append(res, models.MetricRaw{
		Type:  "gauge",
		Name:  "RandomValue",
		Value: m.RandomValue.String(),
	})

	return res
}
