package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/amiskov/metrics-and-alerting/cmd/server/router"
	"github.com/amiskov/metrics-and-alerting/cmd/server/storage"
)

var port string

func init() {
	// CLI options
	serverPort := flag.Int("port", 8080, "server port")
	port = ":" + strconv.Itoa(*serverPort)
}

func main() {
	flag.Parse()
	store := storage.NewServerStore()
	r := router.NewRouter(store)
	fmt.Printf("Server has been started at %s\n", port)
	log.Fatal(http.ListenAndServe(port, r))
}
