package storage

import (
	"github.com/amiskov/metrics-and-alerting/internal/metrics"
)

type MetricData struct {
	MetricName  string
	MetricValue interface{}
}

var gaugeMetrics = make(map[string]metrics.Gauge)
var counterMetrics = make(map[string]metrics.Counter)

func UpdateMetrics(metricData MetricData) {
	switch t := metricData.MetricValue.(type) {
	case metrics.Gauge:
		gaugeMetrics[metricData.MetricName] = t
	case metrics.Counter:
		counterMetrics[metricData.MetricName] += t
	}
}
