package api

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	sm "github.com/amiskov/metrics-and-alerting/cmd/server/models"
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
	rw.Write(jbz)
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
	rw.Write([]byte(`{}`))
}
