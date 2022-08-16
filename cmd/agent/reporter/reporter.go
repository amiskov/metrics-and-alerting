// Package sends metrics over HTTP as JSON or URL params.
package reporter

import (
	"context"
	"log"
	"time"

	"github.com/amiskov/metrics-and-alerting/pkg/models"
)

const (
	withJSON = iota
	withURL
)

type updater interface {
	GetMetrics() []models.Metrics
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

func (a *reporter) ReportWithURLParams() {
	a.runReporter(withURL)
}

func (a *reporter) ReportWithJSON() {
	a.runReporter(withJSON)
}

func (a *reporter) runReporter(apiType int) {
	ticker := time.NewTicker(a.reportInterval)

	go func() {
		<-a.ctx.Done()
		ticker.Stop()
		log.Println("Metrics report stopped.")
		a.done <- true
	}()

	for range ticker.C {
		switch apiType {
		case withJSON:
			a.sendMetricsJSON()
		case withURL:
			a.sendMetrics()
		}
	}
}
