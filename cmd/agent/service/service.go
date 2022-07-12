package service

import (
	"context"
	"fmt"
	"log"
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

func New() *service {
	return &service{
		mx:             new(sync.RWMutex),
		memStats:       new(runtime.MemStats),
		runtimeMetrics: make(map[string]models.Gauge),
	}
}

func (s *service) Run(ctx context.Context, pollInterval time.Duration) {
	ticker := time.NewTicker(pollInterval)
	for range ticker.C {
		select {
		case <-ctx.Done():
			fmt.Println("Cancel update metrics.")
			ticker.Stop()
		default:
			s.updateMetrics()
		}
	}
}

func (s *service) GetMetrics() []models.MetricRaw {
	s.mx.Lock()
	defer s.mx.Unlock()

	var res []models.MetricRaw

	// Get Runtime Metrics
	for name, val := range s.runtimeMetrics {
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
		Value: s.pollCount.String(),
	})

	res = append(res, models.MetricRaw{
		Type:  "gauge",
		Name:  "RandomValue",
		Value: s.randomValue.String(),
	})

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
	s.randomValue = models.Gauge(rand.Float64())
	log.Println("Metrics has been updated.")
}
