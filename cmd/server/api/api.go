package api

import (
	"fmt"
	"log"
	"net/http"

	"github.com/amiskov/metrics-and-alerting/internal/models"
	"github.com/go-chi/chi"
)

type Storage interface {
	UpdateMetric(models.Metrics) error
	GetMetric(string, string) (string, error)
	GetGaugeMetrics() []models.Metrics
	GetCounterMetrics() []models.Metrics
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
