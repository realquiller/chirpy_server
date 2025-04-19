package main

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	server := &http.Server{
		Handler: mux,
		Addr:    ":8080",
	}

	http.HandleFunc("/", MyHandler)
	http.HandleFunc("/healthz", ReadinessHandler)

	log.Fatal(server.ListenAndServe())
}

// go build -o out && ./out
