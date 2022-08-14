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

func (fs *fileStorage) addFileStorage(name string) (func() error, error) {
	file, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE, 0o777)
	if err != nil {
		return nil, err
	}
	fs.file = file
	return file.Close, nil
}

// Decodes JSON from the file and writes it to the given `MetricsDB`.
func (fs *fileStorage) Restore(inmemDB models.InmemDB) error {
	storedMetrics := []models.Metrics{}
	dec := json.NewDecoder(fs.file)
	err := dec.Decode(&storedMetrics)

	if errors.Is(err, io.EOF) {
		log.Println("File is empty, nothing to restore.")
		return nil // empty file is not an error
	}

	if errors.Is(err, nil) {
		for _, m := range storedMetrics {
			updErr := inmemDB.Upsert(m)
			if updErr != nil {
				return updErr
			}
		}
		log.Printf("Metrics data restored from %s", fs.file.Name())
		return nil
	}

	return err
}

func (fs *fileStorage) Dump(inmemDB models.InmemDB) error {
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

	if err := json.NewEncoder(fs.file).Encode(inmemDB.GetAll()); err != nil {
		log.Printf("Can't store to file `%s`: %s.\n", fs.file.Name(), err)
		return err
	}

	log.Printf("Metrics dumped into file `%s`.\n", fs.file.Name())

	return nil
}

func (fs fileStorage) Ping(ctx context.Context) error {
	return nil
}