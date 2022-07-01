package metric

import "time"

type MetricGauge float64
type MetricCounter int64

const defaultPollInterval = 2 * time.Second
const defaultReportInterval = 10 * time.Second

var pollInterval = defaultPollInterval
var reportInterval = defaultReportInterval
