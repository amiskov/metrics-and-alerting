package storage

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/amiskov/metrics-and-alerting/internal/metrics"
)

var incomingMetrics = make(map[string]interface{})

func SaveMetricFromURIPath(path string) error {
	path = strings.TrimPrefix(path, "/update/")
	URIParts := strings.Split(path, "/")

	if len(URIParts) < 3 || len(URIParts) > 3 {
		return errors.New("bad metric")
	}

	metricType := URIParts[0]
	metricName := URIParts[1]
	metricValue := URIParts[2]

	switch metricType {
	case "gauge":
		numVal, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			return err
		}
		incomingMetrics[metricName] = metrics.Gauge(numVal)
	case "counter":
		numVal, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			return err
		}
		incomingMetrics[metricName] = metrics.Counter(numVal)
	default:
		msg := fmt.Sprintf("Unknown metric type %s; %s = %s", metricType, metricName, metricValue)
		return errors.New(msg)
	}

	return nil
}
