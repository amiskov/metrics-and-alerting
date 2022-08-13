package file

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/amiskov/metrics-and-alerting/internal/models"
)

type Cfg struct {
	StoreInterval time.Duration // store immediately if `0`
	Restore       bool          // restore from file on start if `true`
	Ctx           context.Context
	Finished      chan bool
	HashingKey    []byte

	StoreFile string // don't store if `""`
}

type store struct {
	mx            *sync.Mutex
	inmemDB       models.MetricsDB
	storeInterval time.Duration
	file          *os.File
	hashingKey    []byte
}

type closer func() error

func New(cfg *Cfg) (*store, closer, error) {
	shouldUseStoreFile := cfg.StoreFile != ""
	shouldRestoreFromFile := shouldUseStoreFile && cfg.Restore

	s := store{
		mx:            new(sync.Mutex),
		inmemDB:       make(models.MetricsDB),
		storeInterval: cfg.StoreInterval,
		hashingKey:    cfg.HashingKey,
	}

	// Init file Storage
	fileCloser := func() error { return nil }
	if shouldUseStoreFile {
		var err error
		fileCloser, err = s.addFileStorage(cfg.StoreFile)
		if err != nil {
			return nil, nil, err
		}
	}

	// Restore from file or create empty metrics DB
	if shouldUseStoreFile && shouldRestoreFromFile {
		if err := restoreFromFile(s.file, s.inmemDB); err != nil {
			log.Printf("Can't restore from a file %s. Error: %s.", cfg.StoreFile, err)
			return nil, nil, err
		}
	}

	return &s, fileCloser, nil
}

func (s *store) addFileStorage(name string) (func() error, error) {
	file, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE, 0o777)
	if err != nil {
		return nil, err
	}
	s.file = file
	return file.Close, nil
}

func restoreFromFile(file *os.File, metrics models.MetricsDB) error {
	storedMetrics := []models.Metrics{}
	dec := json.NewDecoder(file)
	err := dec.Decode(&storedMetrics)
	switch {
	case errors.Is(err, nil):
		for _, m := range storedMetrics {
			metrics[m.ID] = m
		}
		log.Printf("Metrics data restored from %s", file.Name())
		return nil
	case errors.Is(err, io.EOF):
		log.Println("File is empty, nothing to restore.")
		return nil // empty file is not an error
	default:
		return err
	}
}

func (s *store) Dump() error {
	if _, err := s.file.Stat(); err != nil {
		log.Println("Can't save to file:", err)
		return err
	}

	metrics := s.GetAll()

	if err := s.file.Truncate(0); err != nil {
		log.Println("Can't truncate file contents:", err)
		return err
	}

	if _, err := s.file.Seek(0, 0); err != nil {
		log.Println("Can't set the file offset:", err)
		return err
	}

	if err := json.NewEncoder(s.file).Encode(metrics); err != nil {
		log.Printf("Can't store to file `%s`: %s.\n", s.file.Name(), err)
		return err
	}

	log.Printf("Metrics dumped into file `%s`.\n", s.file.Name())

	return nil
}

// Updates the inmemory metric and stores metrics to file if necessary.
func (s *store) Update(m models.Metrics) error {
	err := s.update(m) // inmemory update
	if err != nil {
		return err
	}

	// Don't dump metrics to file if file is not exists or StoreInterval is not `0`.
	if s.file == nil || s.storeInterval != 0 {
		return nil
	}

	return s.Dump()
}

// Updates the inmem values.
func (s *store) update(m models.Metrics) error {
	s.mx.Lock()
	defer s.mx.Unlock()

	// Metric not exists on the server
	if _, ok := s.inmemDB[m.ID]; !ok {
		s.inmemDB[m.ID] = m
		return nil
	}

	updatedMetric := m

	// Update existing counter metric
	if m.MType == models.MCounter {
		delta := *s.inmemDB[m.ID].Delta + *m.Delta
		updatedMetric.Delta = &delta

		if m.Hash != "" {
			newHash, err := updatedMetric.GetHash(s.hashingKey)
			if err != nil {
				return err
			}
			updatedMetric.Hash = newHash
		}
	}

	s.inmemDB[m.ID] = updatedMetric

	return nil
}

func (s store) GetAll() []models.Metrics {
	s.mx.Lock()
	defer s.mx.Unlock()

	metrics := []models.Metrics{}
	for _, m := range s.inmemDB {
		metrics = append(metrics, m)
	}
	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].ID < metrics[j].ID
	})
	return metrics
}

func (s store) Get(metricType string, metricName string) (models.Metrics, error) {
	s.mx.Lock()
	metric, ok := s.inmemDB[metricName]
	s.mx.Unlock()

	if !ok {
		return metric, models.ErrorMetricNotFound
	}

	return metric, nil
}

func (s store) Ping(ctx context.Context) error {
	return nil
}
