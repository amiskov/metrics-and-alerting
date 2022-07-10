package store

import (
	"math/rand"
	"runtime"
	"sync"

	"github.com/amiskov/metrics-and-alerting/internal/model"
)

type metrics struct {
	mx       *sync.RWMutex
	memStats *runtime.MemStats

	RuntimeMetrics map[string]model.Gauge
	PollCount      model.Counter
	RandomValue    model.Gauge
}

func NewMetricsStore() *metrics {
	return &metrics{
		mx:             new(sync.RWMutex),
		memStats:       new(runtime.MemStats),
		RuntimeMetrics: make(map[string]model.Gauge),
	}
}

func (m *metrics) UpdateAll() {
	runtime.ReadMemStats(m.memStats)

	m.mx.RLock()
	defer m.mx.RUnlock()

	m.RuntimeMetrics["Alloc"] = model.Gauge(m.memStats.Alloc)
	m.RuntimeMetrics["BuckHashSys"] = model.Gauge(m.memStats.BuckHashSys)
	m.RuntimeMetrics["Frees"] = model.Gauge(m.memStats.Frees)
	m.RuntimeMetrics["GCCPUFraction"] = model.Gauge(m.memStats.GCCPUFraction)
	m.RuntimeMetrics["GCSys"] = model.Gauge(m.memStats.GCSys)
	m.RuntimeMetrics["HeapAlloc"] = model.Gauge(m.memStats.HeapAlloc)
	m.RuntimeMetrics["HeapIdle"] = model.Gauge(m.memStats.HeapIdle)
	m.RuntimeMetrics["HeapInuse"] = model.Gauge(m.memStats.HeapInuse)
	m.RuntimeMetrics["HeapObjects"] = model.Gauge(m.memStats.HeapObjects)
	m.RuntimeMetrics["HeapReleased"] = model.Gauge(m.memStats.HeapReleased)
	m.RuntimeMetrics["HeapSys"] = model.Gauge(m.memStats.HeapSys)
	m.RuntimeMetrics["LastGC"] = model.Gauge(m.memStats.LastGC)
	m.RuntimeMetrics["Lookups"] = model.Gauge(m.memStats.Lookups)
	m.RuntimeMetrics["MCacheInuse"] = model.Gauge(m.memStats.MCacheInuse)
	m.RuntimeMetrics["MCacheSys"] = model.Gauge(m.memStats.MCacheSys)
	m.RuntimeMetrics["MSpanInuse"] = model.Gauge(m.memStats.MSpanInuse)
	m.RuntimeMetrics["MSpanSys"] = model.Gauge(m.memStats.MSpanSys)
	m.RuntimeMetrics["Mallocs"] = model.Gauge(m.memStats.Mallocs)
	m.RuntimeMetrics["NextGC"] = model.Gauge(m.memStats.NextGC)
	m.RuntimeMetrics["NumForcedGC"] = model.Gauge(m.memStats.NumForcedGC)
	m.RuntimeMetrics["NumGC"] = model.Gauge(m.memStats.NumGC)
	m.RuntimeMetrics["OtherSys"] = model.Gauge(m.memStats.OtherSys)
	m.RuntimeMetrics["PauseTotalNs"] = model.Gauge(m.memStats.PauseTotalNs)
	m.RuntimeMetrics["StackInuse"] = model.Gauge(m.memStats.StackInuse)
	m.RuntimeMetrics["StackSys"] = model.Gauge(m.memStats.StackSys)
	m.RuntimeMetrics["Sys"] = model.Gauge(m.memStats.Sys)
	m.RuntimeMetrics["TotalAlloc"] = model.Gauge(m.memStats.TotalAlloc)

	m.PollCount++
	m.RandomValue = model.Gauge(rand.Float64())
}

func (m *metrics) GetAll() []model.MetricRaw {
	m.mx.RLock()
	defer m.mx.RUnlock()

	res := []model.MetricRaw{}

	// Get Runtime Metrics
	for name, val := range m.RuntimeMetrics {
		m := model.MetricRaw{
			Type:  "gauge",
			Name:  name,
			Value: val.String(),
		}
		res = append(res, m)
	}

	res = append(res, model.MetricRaw{
		Type:  "counter",
		Name:  "PollCount",
		Value: m.PollCount.String(),
	})

	res = append(res, model.MetricRaw{
		Type:  "gauge",
		Name:  "RandomValue",
		Value: m.RandomValue.String(),
	})

	return res
}
