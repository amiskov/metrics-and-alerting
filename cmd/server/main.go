package main

import (
	"flag"
	"log"
	"net/http"
	"strconv"

	"github.com/amiskov/metrics-and-alerting/cmd/server/storage"
)

var port string

func init() {
	// CLI options
	serverPort := flag.Int("port", 8080, "server port")
	flag.Parse()
	port = ":" + strconv.Itoa(*serverPort)
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		if err := storage.SaveMetricFromURIPath(req.URL.Path); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	})

	log.Fatal(http.ListenAndServe(port, nil))
}
