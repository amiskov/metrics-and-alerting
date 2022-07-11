package main

import (
	"flag"
	"strconv"

	"github.com/amiskov/metrics-and-alerting/cmd/server/api"
	"github.com/amiskov/metrics-and-alerting/cmd/server/store"
)

var port string

func init() {
	// CLI options
	serverPort := flag.Int("port", 8080, "server port")
	port = ":" + strconv.Itoa(*serverPort)
}

func main() {
	flag.Parse()
	storage := store.NewServerStore()
	metricsAPI := api.New(storage)
	metricsAPI.Run(port)
}
