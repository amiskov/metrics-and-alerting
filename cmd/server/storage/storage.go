package storage

import (
	"errors"
	"fmt"
	"sync"

	"github.com/amiskov/metrics-and-alerting/internal/metrics"
)

type MetricData struct {
	MetricName  string
	MetricValue interface{}
}

var gaugeMetrics = make(map[string]metrics.Gauge)
var counterMetrics = make(map[string]metrics.Counter)

var mu sync.Mutex

func UpdateMetrics(metricData MetricData) {
	mu.Lock()
	switch t := metricData.MetricValue.(type) {
	case metrics.Gauge:
		gaugeMetrics[metricData.MetricName] = t
	case metrics.Counter:
		counterMetrics[metricData.MetricName] += t
	}
	mu.Unlock()
}

func GetGaugeMetrics() []string {
	res := []string{}
	mu.Lock()
	for _, gm := range gaugeMetrics {
		res = append(res, gm.String())
	}
	mu.Unlock()
	return res
}

func GetCounterMetrics() []string {
	res := []string{}
	mu.Lock()
	for _, cm := range counterMetrics {
		res = append(res, cm.String())
	}
	mu.Unlock()
	return res
}

var ErrorMetricNotFound = errors.New("metric not found")

func GetMetric(metricType string, metricName string) (string, error) {
	fmt.Println("trying...", metricName)
	switch metricType {
	case "gauge":
		mu.Lock()
		metric, ok := gaugeMetrics[metricName]
		mu.Unlock()
		if !ok {
			return "", ErrorMetricNotFound
		}
		return metric.String(), nil
	case "counter":
		mu.Lock()
		metric, ok := counterMetrics[metricName]
		mu.Unlock()
		if !ok {
			return "", ErrorMetricNotFound
		}
		return metric.String(), nil
	default:
		return "", ErrorMetricNotFound
	}
}
