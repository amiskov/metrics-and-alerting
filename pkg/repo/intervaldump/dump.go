package intervaldump

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/amiskov/metrics-and-alerting/cmd/server/config"
	"github.com/amiskov/metrics-and-alerting/pkg/models"
)

type (
	// `worker` is responsible for interval saving and graceful termination.
	// It intervally dumps data from `source` to `storage` with `ticker`.
	worker struct {
		ctx        context.Context
		terminated chan bool
		ticker     time.Ticker
		source     sourcer
		storage    storer
	}

	// Source of data to dump and destination to restore
	sourcer interface {
		BatchUpsert([]models.Metrics) error
		GetAll() []models.Metrics
	}

	// Persistent storage
	storer interface {
		Restore() ([]models.Metrics, error)
		Dump(context.Context, []models.Metrics) error
	}
)

func New(ctx context.Context, terminated chan bool, source sourcer, storage storer, cfg *config.Config) *worker {
	w := &worker{
		ctx:        ctx,
		terminated: terminated,
		ticker:     *time.NewTicker(cfg.StoreInterval),
		source:     source,
		storage:    storage,
	}

	// Restore metrics from persistent storage or create an empty inmemory DB
	if cfg.Restore {
		err := w.Restore()
		if err != nil {
			log.Println("can't restore metrics from file", cfg.StoreFile)
		}
	}

	return w
}

func (w worker) Run(shouldRestore bool, storeInterval time.Duration) {
	if shouldRestore {
		err := w.Restore()
		if err != nil {
			log.Println("can't restore from a file", err)
		}
		log.Println("restored from file.")
	}

	// Interval saving & restoring from a file
	if storeInterval > 0 {
		go w.DumpPeriodically()
	}

	go w.HandleTermination()
}

func (w worker) Restore() error {
	restoredMetrics, err := w.storage.Restore()
	if err != nil {
		return fmt.Errorf("can't restore data from storage: %w", err)
	}

	errR := w.source.BatchUpsert(restoredMetrics)
	if errR != nil {
		return fmt.Errorf("can't restore data from storage: %w", err)
	}

	return nil
}

// Interval saving to persistent storage.
func (w worker) DumpPeriodically() {
	defer w.ticker.Stop()
	for range w.ticker.C {
		if err := w.storage.Dump(w.ctx, w.source.GetAll()); err != nil {
			log.Println("interval saving to persistent storage failed:", err)
		}
		log.Println("dumped into file")
	}
}

// Handle program termination: save data and stop ticker if it's running.
func (w worker) HandleTermination() {
	<-w.ctx.Done()
	w.ticker.Stop()
	log.Println("Saving timer stopped.")
	if err := w.storage.Dump(w.ctx, w.source.GetAll()); err != nil {
		log.Println("failed saving to persistent storage:", err)
	}
	w.terminated <- true
}
