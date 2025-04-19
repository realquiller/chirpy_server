package handlers

import (
	"net/http"
)

func MyHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Set the Content-Type
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	// 2. Set the status code
	w.WriteHeader(http.StatusOK)

	// 3. Write the response body
	w.Write([]byte("OK"))
}
