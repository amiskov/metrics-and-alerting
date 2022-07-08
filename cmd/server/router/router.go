package router

import (
	"fmt"
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
	fmt.Println(wd)

	if err != nil {
		log.Fatal(err)
	}
	tmpl := template.Must(template.ParseFiles(wd + "/templates/index.html"))

	r.Route("/", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")

			tmpl.Execute(w,
				struct {
					GaugeMetrics   []string
					CounterMetrics []string
				}{
					storage.GetGaugeMetrics(),
					storage.GetCounterMetrics(),
				})
		})
	})

	r.Route("/update", func(r chi.Router) {
		r.Post("/{metricType}/{metricName}/{metricValue}", handlers.UpdateMetricHandler)
	})

	r.Route("/value", func(r chi.Router) {
		r.Get("/{metricType}/{metricName}", func(rw http.ResponseWriter, r *http.Request) {
			metricType := chi.URLParam(r, "metricType")
			metricName := chi.URLParam(r, "metricName")
			metricValue, err := storage.GetMetric(metricType, metricName)
			if err != nil {
				rw.WriteHeader(http.StatusNotFound)
				rw.Write([]byte(err.Error()))
				return
			}
			rw.WriteHeader(http.StatusOK)
			rw.Write([]byte(metricValue))
		})
	})

	return r
}
