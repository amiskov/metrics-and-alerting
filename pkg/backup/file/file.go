package file

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/amiskov/metrics-and-alerting/pkg/models"
)

type fileStorage struct {
	file *os.File
}

type closer func() error

func New(filePath string) (*fileStorage, closer, error) {
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0o777)
	if err != nil {
		return nil, nil, err
	}
	s := fileStorage{
		file: file,
	}
	log.Printf("Using `%s` as a storage.\n", file.Name())
	return &s, file.Close, err
}

// Decodes JSON from the file and writes it to the given `MetricsDB`.
func (fs *fileStorage) ReadAll() ([]models.Metrics, error) {
	storedMetrics := []models.Metrics{}
	dec := json.NewDecoder(fs.file)
	err := dec.Decode(&storedMetrics)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("failed restoring metrics from file `%s`: %w", fs.file.Name(), err)
	}
	return storedMetrics, nil
}

func (fs *fileStorage) Dump(ctx context.Context, metrics []models.Metrics) error {
	if _, err := fs.file.Stat(); err != nil {
		log.Println("Can't save to file:", err)
		return err
	}

	if err := fs.file.Truncate(0); err != nil {
		log.Println("Can't truncate file contents:", err)
		return err
	}

	if _, err := fs.file.Seek(0, 0); err != nil {
		log.Println("Can't set the file offset:", err)
		return err
	}

	if err := json.NewEncoder(fs.file).Encode(metrics); err != nil {
		log.Printf("Can't store to file `%s`: %s.\n", fs.file.Name(), err)
		return err
	}

	log.Printf("Metrics dumped into file `%s`.\n", fs.file.Name())

	return nil
}
