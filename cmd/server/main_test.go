package main_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/amiskov/metrics-and-alerting/cmd/server/handlers"
)

func TestUpdateMetric(t *testing.T) {
	type want struct {
		code        int
		contentType string
	}
	tests := []struct {
		name string
		path string
		want want
	}{
		// Gauge Metrics
		{
			name: "test gauge success",
			path: "gauge/Mallocs/123.00",
			want: want{
				code:        200,
				contentType: "text/plain",
			},
		},
		{
			name: "test counter success",
			path: "counter/PollCount/123",
			want: want{
				code:        http.StatusOK,
				contentType: "text/plain",
			},
		},
		{
			name: "test error metric not provided",
			path: "gauge/",
			want: want{
				code:        http.StatusNotFound,
				contentType: "text/plain",
			},
		},
		{
			name: "test error wrong value type",
			path: "counter/PollCount/0.003",
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain",
			},
		},
		{
			name: "test undefined metric error",
			path: "undefined_metric/must_fail/123",
			want: want{
				code:        http.StatusNotImplemented,
				contentType: "text/plain",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/update/"+tt.path, nil)

			w := httptest.NewRecorder()
			mux := handlers.CreateMux()
			mux.ServeHTTP(w, request)

			res := w.Result()
			defer res.Body.Close()

			if res.StatusCode != tt.want.code {
				t.Errorf("Expected status code %d, got %d", tt.want.code, w.Code)
			}

			if res.Header.Get("Content-Type") != tt.want.contentType {
				t.Errorf("Expected Content-Type %s, got %s", tt.want.contentType, res.Header.Get("Content-Type"))
			}
		})
	}
}
