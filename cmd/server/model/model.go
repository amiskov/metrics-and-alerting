package model

import "errors"

type MetricData struct {
	MetricName  string
	MetricValue interface{}
}

var ErrorMetricNotFound = errors.New("metric not found")
var ErrorBadMetricFormat = errors.New("bad metric format")
var ErrorUnknownMetricType = errors.New("unknown metric type")
