package models_test

import (
	"testing"

	"github.com/amiskov/metrics-and-alerting/pkg/models"
)

func TestMetricHash(t *testing.T) {
	// str := `GetSetZip44:counter:3380724506`
	delta := int64(3380724506)
	m := models.Metrics{
		ID:    "GetSetZip44",
		MType: "counter",
		Delta: &delta,
	}
	key := `/var/folders/x3/8ccpsdrn7cz5q4vnsjj4_rdh0000gn/T/OxSICpv`

	t.Run("test counter hash", func(t *testing.T) {
		expected := "c0979f8d2b6951d1d74fb363eadcf8984cd0e5e2dc2fc5c8ff1eb5d766cf12e8"
		actual, _ := m.GetHash([]byte(key))
		if expected != actual {
			t.Errorf("Expected: %s, got %s", expected, actual)
		}
	})
}
