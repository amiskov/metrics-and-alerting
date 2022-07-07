package storage

import (
	"errors"
	"strconv"
	"strings"

	"github.com/amiskov/metrics-and-alerting/internal/metrics"
)

var incomingMetrics = make(map[string]interface{})
var ErrorWrongMetricType = errors.New("wrong metric type")
var ErrorWrongMetricValue = errors.New("wrong metric value")
var ErrorWrongRequestFormat = errors.New("wrong request format")

func SaveMetricFromURIPath(path string) error {
	path = strings.TrimPrefix(path, "/update/")
	URIParts := strings.Split(path, "/")

	if len(URIParts) < 3 || len(URIParts) > 3 {
		return ErrorWrongRequestFormat
	}

	metricType := URIParts[0]
	metricName := URIParts[1]
	metricValue := URIParts[2]

	switch metricType {
	case "gauge":
		numVal, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			return ErrorWrongMetricValue
		}
		incomingMetrics[metricName] = metrics.Gauge(numVal)
	case "counter":
		numVal, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			return ErrorWrongMetricValue
		}
		incomingMetrics[metricName] = metrics.Counter(numVal)
	default:
		return ErrorWrongMetricType
	}

	return nil
}
