package api

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/amiskov/metrics-and-alerting/internal/models"
)

func (api *metricsAPI) getMetricsListJSON(rw http.ResponseWriter, r *http.Request) {
	jbz, err := json.Marshal(api.store.GetAll())
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		log.Println("Error while parsing all metrics JSON.")
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	writeBody(rw, jbz)
}

func (api *metricsAPI) getMetricJSON(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")

	body, errBody := ioutil.ReadAll(r.Body)
	if errBody != nil {
		log.Println("Error parsing body.", body, errBody)
	}

	var reqMetric models.Metrics
	errj := json.Unmarshal(body, &reqMetric)
	if errj != nil {
		log.Println("Parsing body JSON failed:", errj)
		rw.WriteHeader(http.StatusBadRequest)
		writeBody(rw, []byte(`{"error": "Can't parse body request `+errj.Error()+`"}`))
		return
	}

	foundMetric, err := api.store.Get(reqMetric.MType, reqMetric.ID)
	if err != nil {
		log.Printf("Metric not found. Body: %s. Error: %s.", body, err.Error())
		rw.WriteHeader(http.StatusNotFound)
		writeBody(rw, []byte(`{"error": "Can't get metric `+err.Error()+`"}`))
		return
	}

	jbz, err := json.Marshal(foundMetric)
	if err != nil {
		log.Printf("Error marshaling metric: %+v", err)
		rw.WriteHeader(http.StatusBadRequest)
		writeBody(rw, []byte(`{"error": "`+err.Error()+`"}`))
		return
	}

	rw.WriteHeader(http.StatusOK)
	writeBody(rw, jbz)
}

func (api *metricsAPI) upsertMetricJSON(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")

	body, berr := io.ReadAll(r.Body)
	if berr != nil {
		log.Println("Cant read body")
		return
	}

	metricData := models.Metrics{}
	jErr := json.Unmarshal(body, &metricData)
	if jErr != nil {
		log.Printf("Error while decoding \n`%s`\n error: %v. URL is: %s", body, jErr, r.URL)
		rw.WriteHeader(http.StatusBadRequest)
		writeBody(rw, []byte(`{"error":"`+models.ErrorBadMetricFormat.Error()+`"}`))
		return
	}

	err := api.store.Update(metricData)
	switch {
	case errors.Is(err, models.ErrorBadMetricFormat):
		http.Error(rw, err.Error(), http.StatusBadRequest)
		writeBody(rw, []byte(`{"error":"`+models.ErrorBadMetricFormat.Error()+`"}`))
		return
	case errors.Is(err, models.ErrorMetricNotFound):
		rw.WriteHeader(http.StatusNotFound)
		writeBody(rw, []byte(`{"error":"`+models.ErrorMetricNotFound.Error()+`"}`))
		return
	case errors.Is(err, models.ErrorUnknownMetricType):
		rw.WriteHeader(http.StatusNotImplemented)
		writeBody(rw, []byte(`{"error":"`+models.ErrorUnknownMetricType.Error()+`"}`))
		return
	}

	rw.WriteHeader(http.StatusOK)
	writeBody(rw, []byte(`{}`))
}
