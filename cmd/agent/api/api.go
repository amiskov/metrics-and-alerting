package api

import (
	"context"
	"log"
	"time"
)

type API struct {
	updater        Service
	ctx            context.Context
	done           chan bool
	reportInterval time.Duration
	serverURL      string
}

func New(s Service, ctx context.Context, done chan bool, reportInterval time.Duration, address string) *API {
	return &API{
		updater:        s,
		ctx:            ctx,
		done:           done,
		reportInterval: reportInterval,
		serverURL:      "http://" + address,
	}
}

const (
	withJSON = iota
	withURL
)

func (a *API) ReportWithURLParams() {
	a.runReporter(withURL)
}

func (a *API) ReportWithJSON() {
	a.runReporter(withJSON)
}

func (a *API) runReporter(apiType int) {
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
