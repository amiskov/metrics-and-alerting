package storage

import (
	"errors"
	"sync"

	sm "github.com/amiskov/metrics-and-alerting/cmd/server/model"
	"github.com/amiskov/metrics-and-alerting/internal/model"
)

var ErrorMetricNotFound = errors.New("metric not found")

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

func (s *store) UpdateMetrics(metricData sm.MetricData) {
	s.mx.Lock()
	defer s.mx.Unlock()

	switch t := metricData.MetricValue.(type) {
	case model.Gauge:
		s.GaugeMetrics[metricData.MetricName] = t
	case model.Counter:
		s.CounterMetrics[metricData.MetricName] += t
	}
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
			return "", ErrorMetricNotFound
		}
		return metric.String(), nil
	case "counter":
		metric, ok := s.CounterMetrics[metricName]
		if !ok {
			return "", ErrorMetricNotFound
		}
		return metric.String(), nil
	default:
		return "", ErrorMetricNotFound
	}
}
