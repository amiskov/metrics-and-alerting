// Package sends metrics over HTTP as JSON or URL params.
package reporter

import (
	"context"
	"log"
	"time"

	"github.com/amiskov/metrics-and-alerting/pkg/logger"
	"github.com/amiskov/metrics-and-alerting/pkg/models"
)

const (
	withJSON = iota
	withURL
)

type store interface {
	GetAll() ([]models.Metrics, error)
}

type reporter struct {
	metrics        store
	ctx            context.Context
	terminated     chan bool
	reportInterval time.Duration
	serverURL      string
	hashingKey     []byte
}

func New(ctx context.Context, db store, terminated chan bool,
	reportInterval time.Duration, address string, hashingKey string,
) *reporter {
	return &reporter{
		metrics:        db,
		ctx:            ctx,
		terminated:     terminated,
		reportInterval: reportInterval,
		serverURL:      "http://" + address,
		hashingKey:     []byte(hashingKey),
	}
}

// Run the process which intervally sends metrics from updater as URL params.
func (r *reporter) ReportWithURLParams() {
	r.runReporter(withURL)
}

// Run the process which intervally sends metrics from updater as JSON.
func (r *reporter) ReportWithJSON() {
	r.runReporter(withJSON)
}

func (r *reporter) runReporter(apiType int) {
	ticker := time.NewTicker(r.reportInterval)

	go func() {
		<-r.ctx.Done()
		ticker.Stop()
		log.Println("Metrics report stopped.")
		r.terminated <- true
	}()

	for range ticker.C {
		metrics, err := r.metrics.GetAll()
		if err != nil {
			logger.Log(r.ctx).Errorf("can't get metrics: %v", err)
			return
		}

		// Actualize hashes
		if len(r.hashingKey) > 0 {
			for k, m := range metrics {
				hash, hErr := m.GetHash(r.hashingKey)
				if hErr != nil {
					logger.Log(r.ctx).Error("reporter: failed creating hash %v", hErr)
					return
				}
				metrics[k].Hash = hash
			}
		}

		switch apiType {
		case withJSON:
			r.sendMetricsJSON(metrics)
		case withURL:
			r.sendMetrics(metrics)
		}
	}
}
