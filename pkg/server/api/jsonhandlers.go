package api

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/amiskov/metrics-and-alerting/pkg/logger"
	"github.com/amiskov/metrics-and-alerting/pkg/models"
)

func (api *metricsAPI) getMetricsListJSON(rw http.ResponseWriter, r *http.Request) {
	metrics, err := api.repo.GetAll()
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		logger.Log(r.Context()).Errorf("failed getting metrics: %v", err)
		return
	}

	jbz, err := json.Marshal(metrics)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		logger.Log(r.Context()).Errorf("error while parsing all metrics JSON: %v", err)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	writeBody(r.Context(), rw, jbz)
}

func (api *metricsAPI) getMetricJSON(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")

	body, errRead := ioutil.ReadAll(r.Body)
	if errRead != nil {
		logger.Log(r.Context()).Errorf("error reading request body: %v", errRead)
		return
	}

	var reqMetric models.Metrics
	errJSON := json.Unmarshal(body, &reqMetric)
	if errJSON != nil {
		logger.Log(r.Context()).Errorf("failed parsing request body JSON: %v", errJSON)
		rw.WriteHeader(http.StatusBadRequest)
		writeBody(r.Context(), rw, []byte(`{"error": "Can't parse body request `+errJSON.Error()+`"}`))
		return
	}

	foundMetric, err := api.repo.Get(reqMetric.MType, reqMetric.ID)
	if err != nil {
		logger.Log(r.Context()).Errorf("Metric not found. Body: %s. Error: %v.", body, err)
		rw.WriteHeader(http.StatusNotFound)
		writeBody(r.Context(), rw, []byte(`{"error": "Can't get metric `+err.Error()+`"}`))
		return
	}

	jbz, err := json.Marshal(foundMetric)
	if err != nil {
		logger.Log(r.Context()).Errorf("Error marshaling metric: %+v", err)
		rw.WriteHeader(http.StatusBadRequest)
		writeBody(r.Context(), rw, []byte(`{"error": "`+err.Error()+`"}`))
		return
	}

	rw.WriteHeader(http.StatusOK)
	writeBody(r.Context(), rw, jbz)
}

func (api *metricsAPI) bulkUpdateMetrics(rw http.ResponseWriter, r *http.Request) {
	body, berr := io.ReadAll(r.Body)
	if berr != nil {
		logger.Log(r.Context()).Errorf("can't read request body: %+v", berr)
		return
	}

	metrics := []models.Metrics{}
	jErr := json.Unmarshal(body, &metrics)
	if jErr != nil {
		logger.Log(r.Context()).Errorf("Error while decoding \n`%s`\n error: %v. URL is: %s", body, jErr, r.URL)
		rw.WriteHeader(http.StatusBadRequest)
		writeBody(r.Context(), rw, []byte(`{"error":"`+models.ErrorBadMetricFormat.Error()+`"}`))
		return
	}

	err := api.repo.BulkUpdate(metrics)
	if errors.Is(err, models.ErrorPartialUpdate) {
		logger.Log(r.Context()).Errorf("partial update, some metrics are invalid `%v`", err)
		rw.WriteHeader(http.StatusPartialContent)
		writeBody(r.Context(), rw, []byte(`{"error":"`+err.Error()+`"}`))
		return
	}
	if err != nil {
		logger.Log(r.Context()).Errorf("bulk update failed: %v", body)
		rw.WriteHeader(http.StatusBadRequest)
		writeBody(r.Context(), rw, []byte(`{"error":"`+models.ErrorBadMetricFormat.Error()+`"}`))
		return
	}

	rw.WriteHeader(http.StatusOK)
	writeBody(r.Context(), rw, []byte(`{"message": "metrics updated"}`))
}

func (api *metricsAPI) upsertMetricJSON(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")

	body, berr := io.ReadAll(r.Body)
	if berr != nil {
		logger.Log(r.Context()).Errorf("can't read request body: %v", berr)
		return
	}

	metricData := models.Metrics{}
	jErr := json.Unmarshal(body, &metricData)
	if jErr != nil {
		logger.Log(r.Context()).Errorf("Error while decoding \n`%s`\n error: %v. URL is: %s", body, jErr, r.URL)
		rw.WriteHeader(http.StatusBadRequest)
		writeBody(r.Context(), rw, []byte(`{"error":"`+models.ErrorBadMetricFormat.Error()+`"}`))
		return
	}

	err := api.repo.Update(metricData)
	switch {
	case errors.Is(err, models.ErrorBadMetricFormat):
		http.Error(rw, err.Error(), http.StatusBadRequest)
		writeBody(r.Context(), rw, []byte(`{"error":"`+models.ErrorBadMetricFormat.Error()+`"}`))
		return
	case errors.Is(err, models.ErrorMetricNotFound):
		rw.WriteHeader(http.StatusNotFound)
		writeBody(r.Context(), rw, []byte(`{"error":"`+models.ErrorMetricNotFound.Error()+`"}`))
		return
	case errors.Is(err, models.ErrorUnknownMetricType):
		rw.WriteHeader(http.StatusNotImplemented)
		writeBody(r.Context(), rw, []byte(`{"error":"`+models.ErrorUnknownMetricType.Error()+`"}`))
		return
	}

	rw.WriteHeader(http.StatusOK)
	writeBody(r.Context(), rw, []byte(`{}`))
}
