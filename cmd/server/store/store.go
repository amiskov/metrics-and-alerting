package store

import (
	"sort"
	"sync"

	sm "github.com/amiskov/metrics-and-alerting/cmd/server/models"
	"github.com/amiskov/metrics-and-alerting/internal/models"
)

type store struct {
	mx *sync.Mutex

	GaugeMetrics   map[string]models.Gauge
	CounterMetrics map[string]models.Counter
}

func NewServerStore() *store {
	return &store{
		mx: new(sync.Mutex),

		GaugeMetrics:   make(map[string]models.Gauge),
		CounterMetrics: make(map[string]models.Counter),
	}
}

func (s *store) UpdateMetric(m models.Metrics) error {
	s.mx.Lock()
	defer s.mx.Unlock()

	switch m.MType {
	case "counter":
		s.CounterMetrics[m.ID] += models.Counter(*m.Delta)
	case "gauge":
		s.GaugeMetrics[m.ID] = models.Gauge(*m.Value)
	default:
		return sm.ErrorUnknownMetricType
	}

	return nil
}

func (s store) GetGaugeMetrics() []models.Metrics {
	s.mx.Lock()
	defer s.mx.Unlock()

	var res []models.Metrics

	for name, val := range s.GaugeMetrics {
		val := float64(val)
		res = append(res, models.Metrics{
			MType: "gauge",
			ID:    name,
			Value: &val,
		})
	}

	sort.Slice(res, func(i, j int) bool {
		return res[i].ID < res[j].ID
	})

	return res
}

func (s store) GetCounterMetrics() []models.Metrics {
	s.mx.Lock()
	defer s.mx.Unlock()

	var res []models.Metrics
	for name, val := range s.CounterMetrics {
		val := int64(val)
		res = append(res, models.Metrics{
			MType: "counter",
			ID:    name,
			Delta: &val,
		})
	}

	sort.Slice(res, func(i, j int) bool {
		return res[i].ID < res[j].ID
	})

	return res
}

func (s store) GetMetric(metricType string, metricName string) (string, error) {
	s.mx.Lock()
	defer s.mx.Unlock()

	switch metricType {
	case "gauge":
		metric, ok := s.GaugeMetrics[metricName]
		if !ok {
			return "", sm.ErrorMetricNotFound
		}
		return metric.String(), nil
	case "counter":
		metric, ok := s.CounterMetrics[metricName]
		if !ok {
			return "", sm.ErrorMetricNotFound
		}
		return metric.String(), nil
	default:
		return "", sm.ErrorMetricNotFound
	}
}
