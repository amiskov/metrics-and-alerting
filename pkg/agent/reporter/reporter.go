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

type updater interface {
	GetMetrics() ([]models.Metrics, error)
}

type reporter struct {
	updater        updater
	ctx            context.Context
	done           chan bool
	reportInterval time.Duration
	serverURL      string
}

func New(ctx context.Context, s updater, done chan bool, reportInterval time.Duration, address string) *reporter {
	return &reporter{
		updater:        s,
		ctx:            ctx,
		done:           done,
		reportInterval: reportInterval,
		serverURL:      "http://" + address,
	}
}

func (r *reporter) ReportWithURLParams() {
	r.runReporter(withURL)
}

func (r *reporter) ReportWithJSON() {
	r.runReporter(withJSON)
}

func (r *reporter) runReporter(apiType int) {
	ticker := time.NewTicker(r.reportInterval)

	go func() {
		<-r.ctx.Done()
		ticker.Stop()
		log.Println("Metrics report stopped.")
		r.done <- true
	}()

	for range ticker.C {
		metrics, err := r.updater.GetMetrics()
		if err != nil {
			logger.Log(r.ctx).Errorf("can't get metrics: %v", err)
			return
		}
		switch apiType {
		case withJSON:
			r.sendMetricsJSON(metrics)
		case withURL:
			r.sendMetrics(metrics)
		}
	}
}
