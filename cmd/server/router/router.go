package router

import (
	"html/template"
	"log"
	"net/http"

	sm "github.com/amiskov/metrics-and-alerting/cmd/server/model"
	"github.com/amiskov/metrics-and-alerting/internal/model"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

type Store interface {
	UpdateMetric(model.MetricRaw) error
	GetMetric(string, string) (string, error)
	GetGaugeMetrics() []string
	GetCounterMetrics() []string
}

func NewRouter(s Store) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	tmpl, err := template.New("index").Parse(`<h1>Metrics Service</h1>
		<h2>Gauge Metrics</h2>
		<ul>
		{{range $val := .GaugeMetrics}}
			 <li>{{$val}}</li>
		{{end}}
		</ul>
		<h2>Counter Metrics</h2>
		<ul>
		{{range $val := .CounterMetrics}}
			 <li>{{$val}}</li>
		{{end}}
		</ul>`)
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

			metricData := model.MetricRaw{
				Type:  chi.URLParam(r, "metricType"),
				Name:  chi.URLParam(r, "metricName"),
				Value: chi.URLParam(r, "metricValue"),
			}

			err := s.UpdateMetric(metricData)

			switch err {
			case sm.ErrorBadMetricFormat:
				rw.WriteHeader(http.StatusBadRequest)
				return
			case sm.ErrorBadMetricFormat:
				rw.WriteHeader(http.StatusBadRequest)
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
					s.GetGaugeMetrics(),
					s.GetCounterMetrics(),
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
