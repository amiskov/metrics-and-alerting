package router

import (
	"log"
	"net/http"
	"os"
	"text/template"

	"github.com/amiskov/metrics-and-alerting/cmd/server/handlers"
	"github.com/amiskov/metrics-and-alerting/cmd/server/storage"
	"github.com/go-chi/chi"
)

func NewRouter() chi.Router {
	r := chi.NewRouter()
	wd, err := os.Getwd()

	if err != nil {
		log.Fatal(err)
	}
	tmpl := template.Must(template.ParseFiles(wd + "/templates/index.html"))

	r.Route("/update", func(r chi.Router) {
		r.Post("/{metricType}/", func(rw http.ResponseWriter, r *http.Request) {
			rw.Header().Set("Content-Type", "text/plain")
			rw.WriteHeader(http.StatusNotFound)
		})

		r.Post("/{metricType}/{metricName}/", func(rw http.ResponseWriter, r *http.Request) {
			rw.Header().Set("Content-Type", "text/plain")
			rw.WriteHeader(http.StatusNotImplemented)
		})

		r.Post("/{metricType}/{metricName}/{metricValue}", handlers.UpdateMetricHandler)
	})

	r.Route("/value", func(r chi.Router) {
		r.Get("/{metricType}/{metricName}", func(rw http.ResponseWriter, r *http.Request) {
			metricType := chi.URLParam(r, "metricType")
			metricName := chi.URLParam(r, "metricName")
			metricValue, err := storage.GetMetric(metricType, metricName)
			if err != nil {
				rw.WriteHeader(http.StatusNotFound)
				_, err := rw.Write([]byte(err.Error()))
				if err != nil {
					log.Println("Error writing body")
				}
				return
			}
			rw.WriteHeader(http.StatusOK)
			_, errWriteHeader := rw.Write([]byte(metricValue))
			if errWriteHeader != nil {
				log.Println("Error writing body.")
			}
		})
	})

	r.Route("/", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")

			err := tmpl.Execute(w,
				struct {
					GaugeMetrics   []string
					CounterMetrics []string
				}{
					storage.GetGaugeMetrics(),
					storage.GetCounterMetrics(),
				})
			if err != nil {
				log.Println("error while executing the template")
			}
		})
		r.Post("/*", func(rw http.ResponseWriter, r *http.Request) {
			rw.Header().Set("Content-Type", "text/plain")
			rw.WriteHeader(http.StatusNotFound)
		})
	})

	return r
}
