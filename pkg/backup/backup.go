// Package `backup` can dump metrics intervally from `sourcer` to `storer`.
package backup

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/amiskov/metrics-and-alerting/pkg/logger"
	"github.com/amiskov/metrics-and-alerting/pkg/models"
)

type (
	// `worker` is responsible for interval saving and graceful termination.
	// It dumps metrics intervally from a `sourcer` to a `storer`.
	worker struct {
		ctx        context.Context
		terminated chan bool
		ticker     time.Ticker
		source     Sourcer
		storage    storer
	}

	// Source of data to dump and destination to restore
	Sourcer interface {
		BulkUpdate([]models.Metrics) error
		GetAll() ([]models.Metrics, error)
	}

	// Persistent storage: destination to dump and source to restore
	storer interface {
		ReadAll() ([]models.Metrics, error)
		SaveAll([]models.Metrics) error
	}
)

func New(ctx context.Context, terminated chan bool, source Sourcer, storage storer) *worker {
	return &worker{
		ctx:        ctx,
		terminated: terminated,
		source:     source,
		storage:    storage,
	}
}

func (w worker) Run(shouldRestore bool, storeInterval time.Duration) {
	if shouldRestore {
		err := w.restore()
		if err != nil {
			log.Println("can't restore from a file", err)
		}
		log.Println("restored from file.")
	}

	// Interval saving & restoring from a file
	if storeInterval > 0 {
		w.ticker = *time.NewTicker(storeInterval)
		go w.dumpPeriodically()
	}

	go w.handleTermination()
}

// Runs interval timer for saving metrics to persistent storage.
func (w worker) dumpPeriodically() {
	defer w.ticker.Stop()
	for range w.ticker.C {
		if err := w.dump(); err != nil {
			logger.Log(w.ctx).Errorf("failed saved to file on termination: %v", err)
			return
		}
		log.Println("Successfully saved to file.")
	}
}

// Handles program termination: stops interval saving and dumps the latest metrics snapshot.
func (w worker) handleTermination() {
	<-w.ctx.Done()
	w.ticker.Stop()
	log.Println("Saving timer stopped.")
	if err := w.dump(); err != nil {
		logger.Log(w.ctx).Errorf("failed saved to file on termination: %v", err)
		w.terminated <- true
		return
	}
	log.Println("Successfully saved to file. Terminating.")
	w.terminated <- true
}

// Reads metrics from persistent `storer` and loads into `sourcer`.
func (w worker) restore() error {
	restoredMetrics, err := w.storage.ReadAll()
	if err != nil {
		return fmt.Errorf("can't restore data from storage: %w", err)
	}

	errR := w.source.BulkUpdate(restoredMetrics)
	if errR != nil {
		return fmt.Errorf("can't restore data from storage: %w", err)
	}

	return nil
}

// Saves all data from source to storage
func (w worker) dump() error {
	metrics, err := w.source.GetAll()
	if err != nil {
		return fmt.Errorf("backup: failed getting metrics: %w", err)
	}
	if err := w.storage.SaveAll(metrics); err != nil {
		return fmt.Errorf("backup: failed saving to persistent storage: %w", err)
	}
	return nil
}
