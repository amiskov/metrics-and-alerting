package storage

import (
	"strconv"
	"sync"

	sm "github.com/amiskov/metrics-and-alerting/cmd/server/model"
	"github.com/amiskov/metrics-and-alerting/internal/model"
)

type store struct {
	mx *sync.Mutex

	GaugeMetrics   map[string]model.Gauge
	CounterMetrics map[string]model.Counter
}

func NewServerStore() *store {
	return &store{
		mx: new(sync.Mutex),

		GaugeMetrics:   make(map[string]model.Gauge),
		CounterMetrics: make(map[string]model.Counter),
	}
}

func (s *store) UpdateMetric(m model.MetricRaw) error {
	s.mx.Lock()
	defer s.mx.Unlock()

	switch m.Type {
	case "counter":
		numVal, err := strconv.ParseInt(m.Value, 10, 64)
		if err != nil {
			return sm.ErrorBadMetricFormat
		}
		s.CounterMetrics[m.Name] = model.Counter(numVal)
	case "gauge":
		numVal, err := strconv.ParseFloat(m.Value, 64)
		if err != nil {
			return sm.ErrorBadMetricFormat
		}
		s.GaugeMetrics[m.Name] = model.Gauge(numVal)
	default:
		return sm.ErrorUnknownMetricType
	}

	return nil
}

func (s store) GetGaugeMetrics() []string {
	s.mx.Lock()
	defer s.mx.Unlock()

	res := []string{}

	for _, gm := range s.GaugeMetrics {
		res = append(res, gm.String())
	}
	return res
}

func (s store) GetCounterMetrics() []string {
	s.mx.Lock()
	defer s.mx.Unlock()

	res := []string{}
	for _, cm := range s.CounterMetrics {
		res = append(res, cm.String())
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
