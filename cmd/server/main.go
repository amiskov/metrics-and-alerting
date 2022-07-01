package main

import (
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		// Return 201 after saving
		w.Write([]byte("Got it!"))
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
