package api

import (
	"fmt"
	"log"
	"net/http"

	sm "github.com/amiskov/metrics-and-alerting/cmd/server/models"
	"github.com/amiskov/metrics-and-alerting/internal/models"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

type Storage interface {
	UpdateMetric(models.MetricRaw) error
	GetMetric(string, string) (string, error)
	GetGaugeMetrics() []models.MetricRaw
	GetCounterMetrics() []models.MetricRaw
}

type api struct {
	Router *chi.Mux
	store  Storage
}

func New(s Storage) *api {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	return &api{
		Router: r,
		store:  s,
	}
}

func (a *api) Run(port string) {
	a.Router.Route("/value", func(r chi.Router) {
		r.Get("/{metricType}/{metricName}", func(rw http.ResponseWriter, r *http.Request) {
			metricType := chi.URLParam(r, "metricType")
			metricName := chi.URLParam(r, "metricName")

			metricValue, err := a.store.GetMetric(metricType, metricName)
			if err != nil {
				rw.WriteHeader(http.StatusNotFound)
				rw.Write([]byte(err.Error()))
				return
			}

			rw.WriteHeader(http.StatusOK)
			rw.Write([]byte(metricValue))
		})
	})

	a.Router.Route("/update", func(r chi.Router) {
		r.Post("/{metricType}/", func(rw http.ResponseWriter, r *http.Request) {
			rw.Header().Set("Content-Type", "text/plain")
			rw.WriteHeader(http.StatusNotFound)
		})

		r.Post("/{metricType}/{metricName}/", func(rw http.ResponseWriter, r *http.Request) {
			rw.Header().Set("Content-Type", "text/plain")
			rw.WriteHeader(http.StatusNotImplemented)
		})

		r.Post("/{metricType}/{metricName}/{metricValue}", func(rw http.ResponseWriter, r *http.Request) {
			rw.Header().Set("Content-Type", "text/plain")

			metricData := models.MetricRaw{
				Type:  chi.URLParam(r, "metricType"),
				Name:  chi.URLParam(r, "metricName"),
				Value: chi.URLParam(r, "metricValue"),
			}

			err := a.store.UpdateMetric(metricData)
			switch err {
			case sm.ErrorBadMetricFormat:
				rw.WriteHeader(http.StatusBadRequest)
				return
			case sm.ErrorMetricNotFound:
				rw.WriteHeader(http.StatusNotFound)
				return
			case sm.ErrorUnknownMetricType:
				rw.WriteHeader(http.StatusNotImplemented)
				return
			}

			rw.WriteHeader(http.StatusOK)
		})
	})

	a.Router.Route("/", func(r chi.Router) {
		r.Get("/", a.GetIndex)

		r.Post("/*", func(rw http.ResponseWriter, r *http.Request) {
			rw.Header().Set("Content-Type", "text/plain")
			rw.WriteHeader(http.StatusNotFound)
		})
	})

	fmt.Printf("Server has been started at %s\n", port)
	log.Fatal(http.ListenAndServe(port, a.Router))
}
