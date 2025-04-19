package main

import (
	"log"
	"net/http"

	"github.com/realquiller/chirpy_server/internal/handlers"
)

func main() {
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("/healthz", handlers.ReadinessHandler)

	// App handler
	mux.Handle("/app/", http.StripPrefix("/app/", http.FileServer(http.Dir("./app/"))))

	server := &http.Server{
		Handler: mux,
		Addr:    ":8080",
	}

	log.Fatal(server.ListenAndServe())
}

// go build -o out && ./out
