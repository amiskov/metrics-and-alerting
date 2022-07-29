package api

import (
	"log"
	"net/http"

	"github.com/amiskov/metrics-and-alerting/internal/models"
	"github.com/go-chi/chi"
)

type Storage interface {
	Update(models.Metrics) error
	Get(mType string, mName string) (models.Metrics, error)
	GetAll() []models.Metrics
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
		Addr:    address,
		Handler: api.Router,
	}
	log.Fatalln(server.ListenAndServe())
}
