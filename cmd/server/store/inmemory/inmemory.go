package inmemory

import (
	"sort"
	"sync"

	sm "github.com/amiskov/metrics-and-alerting/cmd/server/models"
	"github.com/amiskov/metrics-and-alerting/internal/models"
)

type inmemoryStore struct {
	mx      *sync.Mutex
	Metrics map[string]models.Metrics
}

func New() *inmemoryStore {
	return &inmemoryStore{
		mx:      new(sync.Mutex),
		Metrics: make(map[string]models.Metrics),
	}
}

func (s *inmemoryStore) Update(m models.Metrics) error {
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

func (s inmemoryStore) GetAll() []models.Metrics {
	var metrics []models.Metrics
	for _, m := range s.Metrics {
		metrics = append(metrics, m)
	}
	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].ID < metrics[j].ID
	})
	return metrics
}

func (s inmemoryStore) Get(metricType string, metricName string) (models.Metrics, error) {
	s.mx.Lock()
	defer s.mx.Unlock()

	metric, ok := s.Metrics[metricName]
	if !ok {
		return metric, sm.ErrorMetricNotFound
	}
	return metric, nil
}
