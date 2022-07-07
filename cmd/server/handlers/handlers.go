package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/amiskov/metrics-and-alerting/cmd/server/storage"
	"github.com/amiskov/metrics-and-alerting/internal/metrics"
)

func CreateMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/update/", UpdateMetricHandler)
	mux.HandleFunc("/", RootHandler)
	return mux
}

func RootHandler(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "text/plain")
	rw.WriteHeader(http.StatusNotFound)
}

var ErrorUnknownMetric = errors.New("unknown metric")
var ErrorMetricNameNotProvided = errors.New("metric name is not provided")
var ErrorMetricValueNotProvided = errors.New("metric value is not provided")
var ErrorMetricBadValue = errors.New("bad metric value format")

func ParsePath(path string) (storage.MetricData, error) {
	// URI format is `/update/<type>/<name>/<value>`
	parts := strings.Split(strings.TrimPrefix(path, "/update/"), "/")
	metricData := storage.MetricData{}

	metricType := parts[0]

	if len(parts) < 2 {
		return metricData, ErrorMetricNameNotProvided
	}
	metricData.MetricName = parts[1]

	if len(parts) < 3 {
		return metricData, ErrorMetricValueNotProvided
	}

	switch metricType {
	case "counter":
		numVal, err := strconv.ParseInt(parts[2], 10, 64)
		if err != nil {
			return metricData, ErrorMetricBadValue
		}
		metricData.MetricValue = metrics.Counter(numVal)
	case "gauge":
		numVal, err := strconv.ParseFloat(parts[2], 64)
		if err != nil {
			return metricData, ErrorMetricBadValue
		}
		metricData.MetricValue = metrics.Counter(numVal)
	default:
		return metricData, ErrorUnknownMetric
	}

	return metricData, nil
}

func UpdateMetricHandler(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "text/plain")
	metricData, err := ParsePath(r.URL.Path)

	switch err {
	case ErrorUnknownMetric:
		rw.WriteHeader(http.StatusNotImplemented)
		return
	case ErrorMetricNameNotProvided:
		rw.WriteHeader(http.StatusNotFound)
		return
	case ErrorMetricValueNotProvided:
		rw.WriteHeader(http.StatusNotFound)
		return
	case ErrorMetricBadValue:
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	storage.UpdateMetrics(metricData)
	rw.WriteHeader(http.StatusOK)
}
