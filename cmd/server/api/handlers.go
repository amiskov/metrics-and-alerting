package api

import (
	"html/template"
	"log"
	"net/http"
	"strconv"

	sm "github.com/amiskov/metrics-and-alerting/cmd/server/models"
	"github.com/amiskov/metrics-and-alerting/internal/models"
	"github.com/go-chi/chi"
)

var indexTmpl = template.Must(
	template.New("index").Parse(`<h1>Metrics Service</h1>
		<h2>Gauge Metrics</h2>
		<table>
		{{range $m := .GaugeMetrics}}
			 <tr><td>{{$m.ID}}</td><td>{{$m.Value}}</td></tr>
		{{end}}
		</table>
		<h2>Counter Metrics</h2>
		<table>
		{{range $m := .CounterMetrics}}
			 <tr><td>{{$m.ID}}</td><td>{{$m.Delta}}</td></tr>
		{{end}}
		</table>`))

func (api *metricsAPI) getMetricsList(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "text/html")

	err := indexTmpl.Execute(rw,
		struct {
			GaugeMetrics   []models.Metrics
			CounterMetrics []models.Metrics
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

	urlVal := chi.URLParam(r, "metricValue")
	mType := chi.URLParam(r, "metricType")
	var val float64
	var delta int64
	switch mType {
	case "counter":
		delta, _ = strconv.ParseInt(urlVal, 10, 64)
	case "gauge":
		val, _ = strconv.ParseFloat(urlVal, 64)
	}
	metricData := models.Metrics{
		MType: mType,
		ID:    chi.URLParam(r, "metricName"),
		Value: &val,
		Delta: &delta,
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
