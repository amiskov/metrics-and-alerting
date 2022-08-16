package api

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func (api *metricsAPI) mountHandlers(l LoggerMiddleware) {
	// add tracing info to request context for better analyzing async call chains
	api.Router.Use(l.SetupTracing)
	// add tracing-aware logger to context
	api.Router.Use(l.SetupLogging)
	// log context dependant request information
	api.Router.Use(l.AccessLog)

	api.Router.Use(middleware.RequestID)
	api.Router.Use(middleware.RealIP)
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
		r.Post("/updates/", api.bulkUpdateMetrics)
		r.Get("/ping", api.ping)
		r.Get("/j", api.getMetricsListJSON)
		r.Post("/*", handleNotFound)
	})
}
