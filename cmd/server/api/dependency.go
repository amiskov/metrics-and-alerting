package api

import "github.com/amiskov/metrics-and-alerting/internal/models"

type Storage interface {
	Update(models.Metrics) error
	Get(mType string, mName string) (models.Metrics, error)
	GetAll() []models.Metrics
}
