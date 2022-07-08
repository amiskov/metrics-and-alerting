package main

import (
	"log"
	"net/http"

	"github.com/amiskov/metrics-and-alerting/cmd/server/router"
)

var port string

// func init() {
// 	// CLI options
// 	serverPort := flag.Int("port", 8080, "server port")
// 	port = ":" + strconv.Itoa(*serverPort)
// }

func main() {
	// flag.Parse()
	router := router.NewRouter()
	// fmt.Printf("Server has been started at %s\n", port)
	log.Fatal(http.ListenAndServe(":8080", router))
}
