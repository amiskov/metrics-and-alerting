package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/amiskov/metrics-and-alerting/cmd/server/router"
)

var port string

func init() {
	// CLI options
	serverPort := flag.Int("port", 8080, "server port")
	port = ":" + strconv.Itoa(*serverPort)
}

func main() {
	flag.Parse()
	router := router.NewRouter()
	fmt.Printf("Server has been started at %s\n", port)
	log.Fatal(http.ListenAndServe(port, router))
}
