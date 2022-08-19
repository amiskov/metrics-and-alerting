package models

import "errors"

var (
	ErrorMetricNotFound    = errors.New("metric not found")
	ErrorBadMetricFormat   = errors.New("bad metric format")
	ErrorPartialUpdate     = errors.New("partial update")
	ErrorUnknownMetricType = errors.New("unknown metric type")
)
