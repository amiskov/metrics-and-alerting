package api

import "github.com/amiskov/metrics-and-alerting/internal/models"

type Service interface {
	GetMetrics() []models.Metrics
}
