package store

import (
	"os"
	"sort"
	"sync"
	"time"

	sm "github.com/amiskov/metrics-and-alerting/cmd/server/models"
	"github.com/amiskov/metrics-and-alerting/internal/models"
)

type StoreCfg struct {
	StoreInterval time.Duration
	StoreFile     string
	Restore       bool
}

type metricsDB map[string]models.Metrics

type store struct {
	mx            *sync.Mutex
	Metrics       metricsDB
	StoreInterval time.Duration
	StoreFile     *os.File
}

func New(cfg StoreCfg) (*store, error) {
	shouldUseStoreFile := cfg.StoreFile != ""
	shouldRestoreFromFile := shouldUseStoreFile && cfg.Restore

	// File Storage
	var file *os.File
	if shouldUseStoreFile {
		file, err := os.OpenFile(cfg.StoreFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)
		defer file.Close()
		if err != nil {
			return nil, err
		}
	}

	// Preload or create metrics DB
	var metrics metricsDB
	if shouldRestoreFromFile {
		metrics = restoreFromFile(file)
	} else {
		metrics = make(metricsDB)
	}

	return &store{
		mx:            new(sync.Mutex),
		StoreInterval: cfg.StoreInterval,
		StoreFile:     file,
		Metrics:       metrics,
	}, nil
}

func restoreFromFile(file *os.File) map[string]models.Metrics {
	// TODO: Load from file
	return make(metricsDB)
}

func (s *store) Update(m models.Metrics) error {
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

func (s store) GetAll() []models.Metrics {
	var metrics []models.Metrics
	for _, m := range s.Metrics {
		metrics = append(metrics, m)
	}
	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].ID < metrics[j].ID
	})
	return metrics
}

func (s store) Get(metricType string, metricName string) (models.Metrics, error) {
	s.mx.Lock()
	defer s.mx.Unlock()

	metric, ok := s.Metrics[metricName]
	if !ok {
		return metric, sm.ErrorMetricNotFound
	}
	return metric, nil
}
