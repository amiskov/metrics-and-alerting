package models

import (
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"fmt"
	"strconv"
)

const (
	MGauge   = "gauge"
	MCounter = "counter"
)

type (
	Gauge   float64
	Counter int64
)

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
	Hash  string   `json:"hash,omitempty"`  // значение хеш-функции
}

func (m Metrics) GetStrVal() (string, error) {
	var val string
	switch m.MType {
	case MGauge:
		val = strconv.FormatFloat(*m.Value, 'f', 3, 64)
	case MCounter:
		val = strconv.FormatInt(*m.Delta, 10)
	default:
		return val, fmt.Errorf("unknown metric type `%s`", m.MType)
	}
	return val, nil
}

func (m Metrics) GetHash(key []byte) (string, error) {
	var src string

	if len(key) == 0 {
		return src, errors.New("hashing key is empty")
	}

	if m.Delta == nil && m.Value == nil {
		return src, errors.New("empty delta and value")
	}

	switch m.MType {
	case MCounter:
		src = fmt.Sprintf("%s:%s:%d", m.ID, m.MType, *m.Delta)
	case MGauge:
		src = fmt.Sprintf("%s:%s:%f", m.ID, m.MType, *m.Value)
	default:
		return src, ErrorUnknownMetricType
	}

	h := hmac.New(sha256.New, key)
	h.Write([]byte(src))

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
