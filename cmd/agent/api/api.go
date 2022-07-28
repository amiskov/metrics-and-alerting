package api

import (
	"context"
	"log"
	"time"

	"github.com/amiskov/metrics-and-alerting/internal/models"
)

const (
	withJSON = iota
	withURL
)

type Service interface {
	GetMetrics() []models.Metrics
}

type api struct {
	updater        Service
	ctx            context.Context
	done           chan bool
	reportInterval time.Duration
	serverURL      string
}

func New(s Service, ctx context.Context, done chan bool, reportInterval time.Duration, address string) *api {
	return &api{
		updater:        s,
		ctx:            ctx,
		done:           done,
		reportInterval: reportInterval,
		serverURL:      "http://" + address,
	}
}

func (a *api) ReportWithURLParams() {
	a.runReporter(withURL)
}

func (a *api) ReportWithJSON() {
	a.runReporter(withJSON)
}

func (a *api) runReporter(apiType int) {
	ticker := time.NewTicker(a.reportInterval)
	for range ticker.C {
		select {
		case <-a.ctx.Done():
			ticker.Stop()
			log.Println("Metrics report stopped.")
			a.done <- true
		default:
			switch apiType {
			case withJSON:
				a.sendMetricsJSON()
			case withURL:
				a.sendMetrics()
			}
		}
	}
}
