package api

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
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
			Metrics: api.store.GetAll(),
		})

	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		log.Println("error while executing the template")
	}
}

func (api *metricsAPI) getMetricsListJSON(rw http.ResponseWriter, r *http.Request) {
	jbz, err := json.Marshal(api.store.GetAll())
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		log.Println("Error while parsing all metrics JSON.")
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	rw.Write(jbz)
}

func (api *metricsAPI) getMetric(rw http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")

	metricValue, err := api.store.Get(metricType, metricName)
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

	body, errBody := ioutil.ReadAll(r.Body)
	if errBody != nil {
		log.Println("Error parsing body.", body, errBody)
	}
	log.Println("This is BODY →", string(body))

	var reqMetric models.Metrics
	errj := json.Unmarshal(body, &reqMetric)
	if errj != nil {
		log.Println("Parsing body JSON failed:", errj)
		return
	}

	foundMetric, err := api.store.Get(reqMetric.MType, reqMetric.ID)
	if err != nil {
		rw.WriteHeader(http.StatusNotFound)
		rw.Write([]byte(`{"error": "Can't get metric ` + err.Error() + `"}`))
		return
	}

	jbz, err := json.Marshal(foundMetric)
	if err != nil {
		log.Printf("Error marshalling metric: %+v", err)
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte(`{"error": "` + err.Error() + `"}`))
		return
	}

	rw.WriteHeader(http.StatusOK)
	log.Println("JSON! →", string(jbz))
	rw.Write(jbz)
}

func (api *metricsAPI) upsertMetric(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "text/plain")

	urlVal := chi.URLParam(r, "metricValue")
	mType := chi.URLParam(r, "metricType")
	var val float64
	var delta int64
	var err error
	switch mType {
	case "counter":
		delta, err = strconv.ParseInt(urlVal, 10, 64)
		if err != nil {
			log.Println(err.Error())
			rw.WriteHeader(http.StatusBadRequest)
			return
		}
	case "gauge":
		val, err = strconv.ParseFloat(urlVal, 64)
		if err != nil {
			log.Println(err.Error())
			rw.WriteHeader(http.StatusBadRequest)
			return
		}
	}
	metricData := models.Metrics{
		MType: mType,
		ID:    chi.URLParam(r, "metricName"),
		Value: &val,
		Delta: &delta,
	}

	err = api.store.Update(metricData)
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
		log.Printf("Error while decoding received metric data: %s. URL is: %s", err, r.URL)
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte(`{"error":"` + sm.ErrorBadMetricFormat.Error() + `"}`))
		return
	}

	err = api.store.Update(metricData)
	switch err {
	case sm.ErrorBadMetricFormat:
		http.Error(rw, err.Error(), http.StatusBadRequest)
		rw.Write([]byte(`{"error":"` + sm.ErrorBadMetricFormat.Error() + `"}`))
		return
	case sm.ErrorMetricNotFound:
		rw.Write([]byte(`{"error":"` + sm.ErrorMetricNotFound.Error() + `"}`))
		rw.WriteHeader(http.StatusNotFound)
		http.NotFound(rw, r)
		return
	case sm.ErrorUnknownMetricType:
		rw.WriteHeader(http.StatusNotImplemented)
		rw.Write([]byte(`{"error":"` + sm.ErrorUnknownMetricType.Error() + `"}`))
		return
	}

	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte(``))
}
