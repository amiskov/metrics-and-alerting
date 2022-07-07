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
		rw.WriteHeader(http.StatusBadRequest)
	} else {
		rw.WriteHeader(http.StatusOK)
	}
}
