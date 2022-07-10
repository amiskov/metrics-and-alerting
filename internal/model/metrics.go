package model

import (
	"strconv"
)

type Gauge float64
type Counter int64

func (g Gauge) String() string {
	return strconv.FormatFloat(float64(g), 'f', 3, 64)
}

func (c Counter) String() string {
	return strconv.FormatInt(int64(c), 10)
}

type MetricForSend struct {
	Type  string
	Name  string
	Value string
}
