package main

import (
	"flag"
	"log"
	"net/http"
	"strconv"
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
		// Return 201 after saving
		log.Println(req.URL.Path)
		w.Write([]byte("Got it!" + req.URL.Path))
	})
	log.Fatal(http.ListenAndServe(port, nil))
}
