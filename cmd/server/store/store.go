package store

import (
	"strconv"
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

func (s *store) UpdateMetric(m models.MetricRaw) error {
	s.mx.Lock()
	defer s.mx.Unlock()

	switch m.Type {
	case "counter":
		numVal, err := strconv.ParseInt(m.Value, 10, 64)
		if err != nil {
			return sm.ErrorBadMetricFormat
		}
		s.CounterMetrics[m.Name] += models.Counter(numVal)
	case "gauge":
		numVal, err := strconv.ParseFloat(m.Value, 64)
		if err != nil {
			return sm.ErrorBadMetricFormat
		}
		s.GaugeMetrics[m.Name] = models.Gauge(numVal)
	default:
		return sm.ErrorUnknownMetricType
	}

	return nil
}

func (s store) GetGaugeMetrics() []models.MetricRaw {
	s.mx.Lock()
	defer s.mx.Unlock()

	res := []models.MetricRaw{}

	for name, val := range s.GaugeMetrics {
		res = append(res, models.MetricRaw{
			Type:  "gauge",
			Name:  name,
			Value: val.String(),
		})
	}
	return res
}

func (s store) GetCounterMetrics() []models.MetricRaw {
	s.mx.Lock()
	defer s.mx.Unlock()

	res := []models.MetricRaw{}
	for name, val := range s.CounterMetrics {
		res = append(res, models.MetricRaw{
			Type:  "counter",
			Name:  name,
			Value: val.String(),
		})
	}
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
