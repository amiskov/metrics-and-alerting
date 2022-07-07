package main

import (
	"flag"
	"log"
	"net/http"
	"strconv"

	"github.com/amiskov/metrics-and-alerting/cmd/server/handlers"
)

var port string

func init() {
	// CLI options
	serverPort := flag.Int("port", 8080, "server port")
	port = ":" + strconv.Itoa(*serverPort)
}

func main() {
	flag.Parse()
	http.HandleFunc("/update/", handlers.UpdateHandler)
	log.Fatal(http.ListenAndServe(port, nil))
}
