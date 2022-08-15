package models

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
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

func (m Metrics) GetHash(key []byte) (string, error) {
	var src string

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
