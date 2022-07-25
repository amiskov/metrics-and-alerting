package store

import (
	"sort"
	"sync"

	sm "github.com/amiskov/metrics-and-alerting/cmd/server/models"
	"github.com/amiskov/metrics-and-alerting/internal/models"
)

type store struct {
	mx      *sync.Mutex
	Metrics map[string]models.Metrics
}

func NewServerStore() *store {
	return &store{
		mx:      new(sync.Mutex),
		Metrics: make(map[string]models.Metrics),
	}
}

func (s *store) UpdateMetric(m models.Metrics) error {
	s.mx.Lock()
	defer s.mx.Unlock()

	switch m.MType {
	case "counter":
		if _, ok := s.Metrics[m.ID]; ok {
			*s.Metrics[m.ID].Delta += *m.Delta
		} else {
			s.Metrics[m.ID] = m
		}
	case "gauge":
		s.Metrics[m.ID] = m
	default:
		return sm.ErrorUnknownMetricType
	}

	return nil
}
func (s store) GetAllMetrics() []models.Metrics {
	var metrics []models.Metrics
	for _, m := range s.Metrics {
		metrics = append(metrics, m)
	}
	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].ID < metrics[j].ID
	})
	return metrics
}

func (s store) GetMetric(metricType string, metricName string) (models.Metrics, error) {
	s.mx.Lock()
	defer s.mx.Unlock()

	metric, ok := s.Metrics[metricName]
	if !ok {
		return metric, sm.ErrorMetricNotFound
	}
	return metric, nil
}
