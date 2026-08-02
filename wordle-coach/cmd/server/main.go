package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/mmrzz/wordle-coach/internal/api"
)

const (
	defaultPort    = "8080"
	defaultOrigins = "http://localhost:5173,http://127.0.0.1:5173"
)

func main() {
	addr := ":" + getenv("PORT", defaultPort)
	origins := parseOrigins(getenv("CORS_ORIGINS", defaultOrigins))

	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", api.Health)

	srv := &http.Server{
		Addr:              addr,
		Handler:           logging(cors(origins)(mux)),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("listening on %s", addr)
	log.Fatal(srv.ListenAndServe())
}

// getenv returns the value of key, or fallback when it is unset or empty.
func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
