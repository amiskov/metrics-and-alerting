package api

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func (api *metricsAPI) mountHandlers() {
	api.Router.Use(middleware.RequestID)
	api.Router.Use(middleware.RealIP)
	api.Router.Use(middleware.Logger)
	api.Router.Use(middleware.Recoverer)
	respTypes := []string{
		"application/javascript",
		"application/json",
		"text/css",
		"text/html",
		"text/plain",
		"text/xml",
	}
	api.Router.Use(middleware.Compress(3, respTypes...))

	api.Router.Route("/value", func(r chi.Router) {
		r.Post("/", api.getMetricJSON)
		r.Get("/{metricType}/{metricName}", api.getMetric)
	})

	api.Router.Route("/update", func(r chi.Router) {
		r.Post("/", api.upsertMetricJSON)
		r.Post("/{metricType}/", handleNotFound)
		r.Post("/{metricType}/{metricName}/", handleNotImplemented)
		r.Post("/{metricType}/{metricName}/{metricValue}", api.upsertMetric)
	})

	api.Router.Route("/", func(r chi.Router) {
		r.Get("/", api.getMetricsList)
		r.Post("/updates/", api.batchUpsertMetrics)
		r.Get("/ping", api.ping)
		r.Get("/j", api.getMetricsListJSON)
		r.Post("/*", handleNotFound)
	})
}
