package handlers

import (
	"log"
	"net/http"

	"github.com/amiskov/metrics-and-alerting/cmd/server/storage"
)

func UpdateHandler(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "text/plain")

	if err := storage.SaveMetricFromURIPath(r.URL.Path); err != nil {
		log.Println(err)
		switch err {
		case storage.ErrorWrongRequestFormat:
			rw.WriteHeader(http.StatusNotImplemented)
		case storage.ErrorWrongMetricType:
			rw.WriteHeader(http.StatusNotImplemented)
		case storage.ErrorWrongMetricValue:
			rw.WriteHeader(http.StatusNotImplemented)
		default:
			rw.WriteHeader(http.StatusNotFound)
		}
	} else {
		rw.WriteHeader(http.StatusOK)
	}
}
