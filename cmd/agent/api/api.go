package api

import (
	"context"
	"log"
	"time"

	"github.com/amiskov/metrics-and-alerting/internal/models"
)

type Service interface {
	GetMetrics() []models.Metrics
}

type api struct {
	service Service
}

func New(s Service) *api {
	return &api{service: s}
}

func (a *api) Run(ctx context.Context, done chan bool, reportInterval time.Duration, serverURL string) {
	ticker := time.NewTicker(reportInterval)
	for range ticker.C {
		select {
		case <-ctx.Done():
			ticker.Stop()
			log.Println("Metrics report stopped.")
			done <- true
		default:
			a.sendMetrics(serverURL)
		}
	}
}

func (a *api) RunJSON(ctx context.Context, done chan bool, reportInterval time.Duration, serverURL string) {
	ticker := time.NewTicker(reportInterval)
	for range ticker.C {
		select {
		case <-ctx.Done():
			ticker.Stop()
			log.Println("Metrics report stopped.")
			done <- true
		default:
			a.sendMetricsJSON(serverURL)
		}
	}
}
