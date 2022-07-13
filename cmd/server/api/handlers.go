package api

import (
	sm "github.com/amiskov/metrics-and-alerting/cmd/server/models"
	"github.com/amiskov/metrics-and-alerting/internal/models"
	"github.com/go-chi/chi"
	"html/template"
	"log"
	"net/http"
)

var indexTmpl = template.Must(
	template.New("index").Parse(`<h1>Metrics Service</h1>
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
		</table>`))

func (api *metricsAPI) getMetricsList(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "text/html")

	err := indexTmpl.Execute(rw,
		struct {
			GaugeMetrics   []models.MetricRaw
			CounterMetrics []models.MetricRaw
		}{
			api.store.GetGaugeMetrics(),
			api.store.GetCounterMetrics(),
		})

	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		log.Println("error while executing the template")
	}
}

func (api *metricsAPI) getMetric(rw http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")

	metricValue, err := api.store.GetMetric(metricType, metricName)
	if err != nil {
		rw.WriteHeader(http.StatusNotFound)
		rw.Write([]byte(err.Error()))
		return
	}

	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte(metricValue))
}

func (api *metricsAPI) upsertMetric(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "text/plain")

	metricData := models.MetricRaw{
		Type:  chi.URLParam(r, "metricType"),
		Name:  chi.URLParam(r, "metricName"),
		Value: chi.URLParam(r, "metricValue"),
	}

	err := api.store.UpdateMetric(metricData)
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
}

func handleNotFound(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "text/plain")
	rw.WriteHeader(http.StatusNotFound)
}

func handleNotImplemented(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "text/plain")
	rw.WriteHeader(http.StatusNotImplemented)
}
