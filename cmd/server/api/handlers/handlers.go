package handlers

import (
	"net/http"
	"strconv"

	"github.com/amiskov/metrics-and-alerting/cmd/server/storage"
	"github.com/amiskov/metrics-and-alerting/internal/model"
	"github.com/go-chi/chi"
)

func UpdateMetricHandler(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "text/plain")

	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")
	metricValue := chi.URLParam(r, "metricValue")

	metricData := storage.MetricData{
		MetricName: metricName,
	}

	switch metricType {
	case "counter":
		numVal, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			return
		}
		metricData.MetricValue = model.Counter(numVal)
	case "gauge":
		numVal, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			return
		}
		metricData.MetricValue = model.Gauge(numVal)
	default:
		rw.WriteHeader(http.StatusNotImplemented)
		return
	}

	storage.UpdateMetrics(metricData)
	rw.WriteHeader(http.StatusOK)
}
