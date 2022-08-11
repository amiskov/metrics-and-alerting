package main_test

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/amiskov/metrics-and-alerting/cmd/server/api"
	"github.com/amiskov/metrics-and-alerting/cmd/server/store"
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
			path: "/update/gauge/Mallocs/123.00",
			want: want{
				code:        200,
				contentType: "text/plain",
			},
		},
		{
			name: "test counter success",
			path: "/update/counter/PollCount/123",
			want: want{
				code:        http.StatusOK,
				contentType: "text/plain",
			},
		},
		{
			name: "test error metric not provided",
			path: "/update/gauge/",
			want: want{
				code:        http.StatusNotFound,
				contentType: "text/plain",
			},
		},
		{
			name: "test error wrong value type",
			path: "/update/counter/PollCount/0.003",
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain",
			},
		},
		{
			name: "test undefined metric error",
			path: "/update/undefined_metric/must_fail/123",
			want: want{
				code:        http.StatusNotImplemented,
				contentType: "text/plain",
			},
		},
		{
			name: "test not existing route",
			path: "/updateme/counter/test/123",
			want: want{
				code:        http.StatusNotFound,
				contentType: "text/plain",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, tt.path, nil)

			w := httptest.NewRecorder()

			ctx, cancel := context.WithCancel(context.Background())
			finished := make(chan bool)
			defer cancel()

			storage, closeFile, err := store.New(&store.Cfg{
				Ctx:       ctx,
				Finished:  finished,
				StoreFile: "",
				Restore:   false,
			})
			if err != nil {
				log.Fatalln("Creating server store failed.", err)
			}
			defer func() {
				if err := closeFile(); err != nil {
					log.Println("failed closing file storage:", err)
				}
			}()
			metricsAPI := api.New(storage)
			metricsAPI.Router.ServeHTTP(w, request)

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
