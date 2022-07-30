package store

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

	sm "github.com/amiskov/metrics-and-alerting/cmd/server/models"
	"github.com/amiskov/metrics-and-alerting/internal/models"
)

type Cfg struct {
	StoreInterval time.Duration // store immediately if `0`
	StoreFile     string        // don't store if `""`
	Restore       bool          // restore from file on start if `true`
	Ctx           context.Context
	Finished      chan bool
}

type metricsDB map[string]models.Metrics

type store struct {
	mx            *sync.Mutex
	metrics       metricsDB
	storeInterval time.Duration
	file          *os.File
}

func New(cfg *Cfg) (*store, func(), error) {
	shouldUseStoreFile := cfg.StoreFile != ""
	shouldRestoreFromFile := shouldUseStoreFile && cfg.Restore

	s := store{
		mx:            new(sync.Mutex),
		metrics:       make(metricsDB),
		storeInterval: cfg.StoreInterval,
		file:          new(os.File),
	}

	var fileCloser func()

	// File Storage
	if shouldUseStoreFile {
		closeFile, err := s.addFileStorage(cfg.StoreFile)
		if err != nil {
			return nil, nil, err
		}
		fileCloser = func() {
			if err := closeFile(); err != nil {
				log.Fatalln("Failed while closing StoreFile: ", err)
			}
		}
	}

	// Restore from file or create empty metrics DB
	if shouldRestoreFromFile {
		if err := restoreFromFile(s.file, s.metrics); err != nil {
			log.Printf("Can't restore from a file %s. Error: %s.", cfg.StoreFile, err)
			return nil, nil, err
		}
		log.Printf("Metrics data restored from %s", cfg.StoreFile)
	}

	save := func() {
		if err := s.saveToFile(); err != nil {
			log.Println("Failed saving metrics to file.", err)
			return
		}
		log.Printf("Metrics successfully saved to `%s`.", s.file.Name())
	}

	var ticker *time.Ticker

	go func() {
		if s.storeInterval != 0 {
			ticker = time.NewTicker(s.storeInterval)
			defer ticker.Stop()
			for range ticker.C {
				save()
			}
		}
	}()

	go func() {
		<-cfg.Ctx.Done()
		if ticker != nil {
			ticker.Stop()
		}
		log.Println("Saving timer stopped.")
		save()
		cfg.Finished <- true
	}()

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

func restoreFromFile(file *os.File, metrics metricsDB) error {
	storedMetrics := []models.Metrics{}
	dec := json.NewDecoder(file)
	err := dec.Decode(&storedMetrics)
	switch {
	case errors.Is(err, nil):
		for _, m := range storedMetrics {
			metrics[m.ID] = m
		}
		return nil
	case errors.Is(err, io.EOF):
		log.Println("File is empty, nothing to restore.")
		return nil // empty file is not an error
	default:
		return err
	}
}

func (s *store) saveToFile() error {
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
	return nil
}

func (s *store) CloseFile() error {
	return s.file.Close()
}

func (s *store) Update(m models.Metrics) error {
	s.mx.Lock()
	defer s.mx.Unlock()

	switch m.MType {
	case models.MCounter:
		if _, ok := s.metrics[m.ID]; ok {
			*s.metrics[m.ID].Delta += *m.Delta
		} else {
			s.metrics[m.ID] = m
		}
	case models.MGauge:
		s.metrics[m.ID] = m
	default:
		return sm.ErrorUnknownMetricType
	}

	// Save to disk immediately
	if s.storeInterval == 0 {
		if err := s.saveToFile(); err != nil {
			log.Println("Can't save metrics to file on Update:", err)
		}
	}

	return nil
}

func (s store) GetAll() []models.Metrics {
	s.mx.Lock()
	defer s.mx.Unlock()

	metrics := []models.Metrics{}
	for _, m := range s.metrics {
		metrics = append(metrics, m)
	}
	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].ID < metrics[j].ID
	})
	return metrics
}

func (s store) Get(metricType string, metricName string) (models.Metrics, error) {
	if metricType != models.MCounter && metricType != models.MGauge {
		return models.Metrics{}, sm.ErrorMetricNotFound
	}

	s.mx.Lock()
	metric, ok := s.metrics[metricName]
	s.mx.Unlock()

	if !ok {
		return metric, sm.ErrorMetricNotFound
	}
	return metric, nil
}
