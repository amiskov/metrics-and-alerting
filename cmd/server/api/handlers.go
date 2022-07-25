package api

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strconv"

	sm "github.com/amiskov/metrics-and-alerting/cmd/server/models"
	"github.com/amiskov/metrics-and-alerting/internal/models"
	"github.com/go-chi/chi"
)

var indexTmpl = template.Must(
	template.New("index").Parse(`<h1>Metrics</h1>
		<table>
		{{range $m := .Metrics}}
			 <tr>
			 <td>{{$m.ID}}</td>
			 	<td>{{$m.MType}}</td>

			 {{if (eq $m.MType "gauge")}}
			 	<td>{{$m.Value}}</td>
			 {{ end }}

			 {{if (eq $m.MType "counter")}}
			 	<td>{{$m.Delta}}</td>
			 {{ end }}
			 </tr>
		{{end}}
		</table>`))

func (api *metricsAPI) getMetricsList(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "text/html")

	err := indexTmpl.Execute(rw,
		struct {
			Metrics []models.Metrics
		}{
			Metrics: api.store.GetAllMetrics(),
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
	var res string
	switch metricType {
	case "counter":
		res = strconv.FormatInt(*metricValue.Delta, 10)
	case "gauge":
		res = strconv.FormatFloat(*metricValue.Value, 'f', 3, 64)
	}

	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte(res))
}

func (api *metricsAPI) getMetricJSON(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")

	reqMetric := models.Metrics{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&reqMetric)
	if err != nil {
		log.Printf("Error while decoding metric for the response: %s", err)
	}
	metricType := reqMetric.MType
	metricName := reqMetric.ID

	metricValue, err := api.store.GetMetric(metricType, metricName)
	if err != nil {
		rw.WriteHeader(http.StatusNotFound)
		rw.Write([]byte(`{"error": "` + err.Error() + `"}`))
		return
	}

	jbz, err := json.Marshal(metricValue)
	if err != nil {
		log.Printf("Error marshalling JSON: %+v", err)
	}

	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte(string(jbz)))
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

func (api *metricsAPI) upsertMetricJSON(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")

	metricData := models.Metrics{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&metricData)
	if err != nil {
		log.Printf("Error while decoding received metric data: %s. Body is: %#v", err, r.Body)
	}

	err = api.store.UpdateMetric(metricData)
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
