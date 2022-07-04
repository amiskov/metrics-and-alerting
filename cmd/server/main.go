package main

import (
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		// Return 201 after saving
		log.Println(req.URL.Path)
		w.Write([]byte("Got it!" + req.URL.Path))
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
