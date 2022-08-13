package file

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"

	"github.com/amiskov/metrics-and-alerting/internal/models"
)

type fileStorage struct {
	file *os.File
}

type closer func() error

func New(filePath string) (*fileStorage, closer, error) {
	shouldUseStoreFile := filePath != ""

	s := fileStorage{}

	// Init file Storage
	fileCloser := func() error { return nil }
	if shouldUseStoreFile {
		var err error
		fileCloser, err = s.addFileStorage(filePath)
		if err != nil {
			return nil, nil, err
		}
	}

	return &s, fileCloser, nil
}

func (s *fileStorage) addFileStorage(name string) (func() error, error) {
	file, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE, 0o777)
	if err != nil {
		return nil, err
	}
	s.file = file
	return file.Close, nil
}

func (s *fileStorage) Restore(inmemDB models.MetricsDB) error {
	storedMetrics := []models.Metrics{}
	dec := json.NewDecoder(s.file)
	err := dec.Decode(&storedMetrics)

	if errors.Is(err, io.EOF) {
		log.Println("File is empty, nothing to restore.")
		return nil // empty file is not an error
	}

	if errors.Is(err, nil) {
		for _, m := range storedMetrics {
			inmemDB[m.ID] = m
		}
		log.Printf("Metrics data restored from %s", s.file.Name())
		return nil
	}

	return err
}

func (s *fileStorage) Dump(inmemDB models.MetricsDB) error {
	if _, err := s.file.Stat(); err != nil {
		log.Println("Can't save to file:", err)
		return err
	}

	if err := s.file.Truncate(0); err != nil {
		log.Println("Can't truncate file contents:", err)
		return err
	}

	if _, err := s.file.Seek(0, 0); err != nil {
		log.Println("Can't set the file offset:", err)
		return err
	}

	if err := json.NewEncoder(s.file).Encode(inmemDB); err != nil {
		log.Printf("Can't store to file `%s`: %s.\n", s.file.Name(), err)
		return err
	}

	log.Printf("Metrics dumped into file `%s`.\n", s.file.Name())

	return nil
}

func (s fileStorage) Ping(ctx context.Context) error {
	return nil
}
