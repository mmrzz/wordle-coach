package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/mmrzz/wordle-coach/internal/api"
	"github.com/mmrzz/wordle-coach/internal/data"
	"github.com/mmrzz/wordle-coach/internal/solver"
)

const (
	defaultPort    = "8080"
	defaultOrigins = "http://localhost:5173,http://127.0.0.1:5173"
)

func main() {
	addr := ":" + getenv("PORT", defaultPort)
	origins := parseOrigins(getenv("CORS_ORIGINS", defaultOrigins))

	// Loading up front turns a corrupt word list into a startup failure rather
	// than a 500 on the first request.
	set, err := data.Load()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("loaded %d answers, %d allowed guesses", len(set.Answers), len(set.Allowed))

	// The engine holds no per-game state, so one instance serves every request.
	solve := api.NewSolver(solver.New(set))

	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", api.Health)
	mux.HandleFunc("GET /api/words", api.NewWords(set).List)
	mux.HandleFunc("POST /api/suggest", solve.Suggest)
	mux.HandleFunc("POST /api/rate", solve.Rate)

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
