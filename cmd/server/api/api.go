package api

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi"

	"github.com/amiskov/metrics-and-alerting/internal/models"
)

type Repo interface {
	Ping(context.Context) error
	Get(metricType string, metricName string) (models.Metrics, error)
	GetAll() []models.Metrics
	Update(m models.Metrics) error
}

type metricsAPI struct {
	Router *chi.Mux
	repo   Repo
}

func New(s Repo) *metricsAPI {
	api := &metricsAPI{
		Router: chi.NewRouter(),
		repo:   s,
	}
	api.mountHandlers()
	return api
}

func (api *metricsAPI) Run(address string) {
	server := &http.Server{
		Addr:              address,
		Handler:           api.Router,
		ReadHeaderTimeout: 2 * time.Second,
	}
	log.Fatalln(server.ListenAndServe())
}
