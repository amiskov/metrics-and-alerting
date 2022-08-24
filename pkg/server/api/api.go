package api

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi"

	"github.com/amiskov/metrics-and-alerting/pkg/models"
)

type Repo interface {
	Ping(context.Context) error
	Get(metricType string, metricName string) (models.Metrics, error)
	GetAll() ([]models.Metrics, error)
	Update(models.Metrics) error
	BulkUpdate([]models.Metrics) (int, error)
}

type metricsAPI struct {
	Router *chi.Mux
	repo   Repo
}

type LoggerMiddleware interface {
	SetupTracing(http.Handler) http.Handler
	SetupLogging(http.Handler) http.Handler
	AccessLog(http.Handler) http.Handler
}

func New(s Repo, l LoggerMiddleware) *metricsAPI {
	api := &metricsAPI{
		Router: chi.NewRouter(),
		repo:   s,
	}
	api.mountHandlers(l)
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
