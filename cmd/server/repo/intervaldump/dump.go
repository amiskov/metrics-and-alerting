package intervaldump

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/amiskov/metrics-and-alerting/cmd/server/config"
	"github.com/amiskov/metrics-and-alerting/internal/models"
)

type storer interface {
	Restore() ([]models.Metrics, error)
	Dump(context.Context, []models.Metrics) error
}

type sourcer interface {
	BatchUpsert([]models.Metrics) error
	GetAll() []models.Metrics
}

type worker struct {
	Ctx        context.Context
	Finished   chan bool
	ticker     time.Ticker
	storage    sourcer
	dumper     storer
	hashingKey []byte
}

func New(ctx context.Context, finished chan bool, storage sourcer, dumper storer, cfg *config.Config) *worker {
	id := &worker{
		Ctx:        ctx,
		Finished:   finished,
		ticker:     *time.NewTicker(cfg.StoreInterval),
		storage:    storage,
		dumper:     dumper,
		hashingKey: []byte(cfg.HashingKey),
	}

	// Restore metrics from persistent storage or create an empty inmemory DB
	if cfg.Restore {
		err := id.Restore()
		if err != nil {
			log.Println("can't restore metrics from file", cfg.StoreFile)
		}
	}

	return id
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

	go w.HandleTermination() // ctx, finished
}

func (w worker) Restore() error {
	restoredMetrics, err := w.dumper.Restore()
	if err != nil {
		return fmt.Errorf("can't restore data from storage: %w", err)
	}

	errR := w.storage.BatchUpsert(restoredMetrics)
	if errR != nil {
		return fmt.Errorf("can't restore data from storage: %w", err)
	}

	return nil
}

// Interval saving to persistent storage.
func (w worker) DumpPeriodically() {
	defer w.ticker.Stop()
	for range w.ticker.C {
		if err := w.dumper.Dump(w.Ctx, w.storage.GetAll()); err != nil {
			log.Println("interval saving to persistent storage failed:", err)
		}
		log.Println("dumped into file")
	}
}

// Handle program termination: save data and stop ticker if it's running.
func (w worker) HandleTermination() {
	<-w.Ctx.Done()
	w.ticker.Stop()
	log.Println("Saving timer stopped.")
	if err := w.dumper.Dump(w.Ctx, w.storage.GetAll()); err != nil {
		log.Println("failed saving to persistent storage:", err)
	}
	w.Finished <- true
}
