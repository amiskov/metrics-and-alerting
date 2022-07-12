package api

import (
	"fmt"
	"github.com/go-chi/chi/middleware"
	"log"
	"net/http"

	"github.com/amiskov/metrics-and-alerting/internal/models"
	"github.com/go-chi/chi"
)

type Storage interface {
	UpdateMetric(models.MetricRaw) error
	GetMetric(string, string) (string, error)
	GetGaugeMetrics() []models.MetricRaw
	GetCounterMetrics() []models.MetricRaw
}

type metricsAPI struct {
	Router *chi.Mux
	store  Storage
}

func New(s Storage) *metricsAPI {
	api := &metricsAPI{
		Router: chi.NewRouter(),
		store:  s,
	}
	api.mountHandlers()
	return api
}

func (api *metricsAPI) Run(port string) {
	fmt.Printf("Server has been started at %s\n", port)
	log.Fatal(http.ListenAndServe(port, api.Router))
}
func (api *metricsAPI) MountHndlrs(handler *chi.Mux) {
	api.Router = handler
}

func (api *metricsAPI) mountHandlers() {
	api.Router.Use(middleware.RequestID)
	api.Router.Use(middleware.RealIP)
	api.Router.Use(middleware.Logger)
	api.Router.Use(middleware.Recoverer)

	api.Router.Route("/value", func(r chi.Router) {
		r.Get("/{metricType}/{metricName}", api.getMetric)
	})

	api.Router.Route("/update", func(r chi.Router) {
		r.Post("/{metricType}/", handleNotFound)
		r.Post("/{metricType}/{metricName}/", handleNotImplemented)
		r.Post("/{metricType}/{metricName}/{metricValue}", api.upsertMetric)
	})

	api.Router.Route("/", func(r chi.Router) {
		r.Get("/", api.getMetricsList)
		r.Post("/*", handleNotFound)
	})
}
