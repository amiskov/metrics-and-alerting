package models

import "errors"

var ErrorMetricNotFound = errors.New("metric not found")
var ErrorBadMetricFormat = errors.New("bad metric format")
var ErrorUnknownMetricType = errors.New("unknown metric type")
