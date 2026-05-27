package main

import (
	"log"
	"net/http"
	"os"

	"endpoint-parser/internal/server"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	h := server.NewHandler()

	mux := http.NewServeMux()
	mux.HandleFunc("/parse", h.HandleParse)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	log.Printf("endpoint-parser listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
