package api

import (
	"html/template"
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

func NewRouter(s Storage) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	tmpl, err := template.New("index").Parse(`<h1>Metrics Service</h1>
		<h2>Gauge Metrics</h2>
		<table>
		{{range $m := .GaugeMetrics}}
			 <tr><td>{{$m.Name}}</td><td>{{$m.Value}}</td></tr>
		{{end}}
		</table>
		<h2>Counter Metrics</h2>
		<table>
		{{range $m := .CounterMetrics}}
			 <tr><td>{{$m.Name}}</td><td>{{$m.Value}}</td></tr>
		{{end}}
		</table>`)
	if err != nil {
		log.Fatal(err)
	}

	r.Route("/update", func(r chi.Router) {
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

			err := s.UpdateMetric(metricData)
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

	r.Route("/value", func(r chi.Router) {
		r.Get("/{metricType}/{metricName}", func(rw http.ResponseWriter, r *http.Request) {
			metricType := chi.URLParam(r, "metricType")
			metricName := chi.URLParam(r, "metricName")

			metricValue, err := s.GetMetric(metricType, metricName)
			if err != nil {
				rw.WriteHeader(http.StatusNotFound)
				rw.Write([]byte(err.Error()))
				return
			}

			rw.WriteHeader(http.StatusOK)
			rw.Write([]byte(metricValue))
		})
	})

	r.Route("/", func(r chi.Router) {
		r.Get("/", func(rw http.ResponseWriter, r *http.Request) {
			rw.Header().Set("Content-Type", "text/html")

			err := tmpl.Execute(rw,
				struct {
					GaugeMetrics   []models.MetricRaw
					CounterMetrics []models.MetricRaw
				}{
					s.GetGaugeMetrics(),
					s.GetCounterMetrics(),
				})
			if err != nil {
				rw.WriteHeader(http.StatusInternalServerError)
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
