package api

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi"

	"github.com/amiskov/metrics-and-alerting/internal/models"
)

type Storage interface {
	Update(models.Metrics) error
	Get(mType string, mName string) (models.Metrics, error)
	GetAll() []models.Metrics
	Ping(context.Context) error
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

func (api *metricsAPI) Run(address string) {
	server := &http.Server{
		Addr:              address,
		Handler:           api.Router,
		ReadHeaderTimeout: 2 * time.Second,
	}
	log.Fatalln(server.ListenAndServe())
}
