package common

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"

	sm "github.com/amiskov/metrics-and-alerting/cmd/server/models"
	"github.com/amiskov/metrics-and-alerting/internal/models"
)

func Hash(m models.Metrics, key []byte) (string, error) {
	var src string

	switch m.MType {
	case models.MCounter:
		src = fmt.Sprintf("%s:%s:%d", m.ID, models.MCounter, *m.Delta)
	case models.MGauge:
		src = fmt.Sprintf("%s:%s:%f", m.ID, models.MGauge, *m.Value)
	default:
		return src, sm.ErrorUnknownMetricType
	}

	h := hmac.New(sha256.New, key)
	h.Write([]byte(src))
	dst := h.Sum(nil)

	return fmt.Sprintf("%x", dst), nil
}
