package repo

import (
	"context"
	"time"

	"github.com/amiskov/metrics-and-alerting/cmd/server/config"
	"github.com/amiskov/metrics-and-alerting/internal/models"
)

type Storage interface {
	Ping(context.Context) error
	Get(metricType string, metricName string) (models.Metrics, error)
	GetAll() []models.Metrics
	Update(m models.Metrics) error
}

type Repo struct {
	StoreInterval time.Duration // store immediately if `0`
	Restore       bool          // restore from file on start if `true`
	Ctx           context.Context
	Finished      chan bool
	HashingKey    []byte
	DB            Storage
}

func New(cfg *config.Config, s Storage) *Repo {
	return &Repo{
		DB: s,
	}
}

func (r Repo) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return r.DB.Ping(ctx)
}

func (r Repo) Get(metricType string, metricName string) (models.Metrics, error) {
	return r.DB.Get(metricType, metricName)
}

func (r Repo) GetAll() []models.Metrics {
	return r.DB.GetAll()
}

func (r *Repo) Update(m models.Metrics) error {
	return r.DB.Update(m)
}
