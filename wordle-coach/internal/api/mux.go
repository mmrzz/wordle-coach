package api

import (
	"net/http"

	"github.com/mmrzz/wordle-coach/internal/data"
	"github.com/mmrzz/wordle-coach/internal/solver"
)

// NewMux wires every route the coach serves.
//
// The routing table lives here rather than in a main, because the same set of
// routes has to answer whether they are served by a long-running process or by
// a function that is handed one request at a time. Anything specific to how it
// is hosted — CORS, logging, listening at all — stays outside.
func NewMux(set *data.Set) *http.ServeMux {
	// The engine holds no per-game state, so one instance serves every request.
	solve := NewSolver(solver.New(set))

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", Health)
	mux.HandleFunc("GET /api/words", NewWords(set).List)
	mux.HandleFunc("POST /api/suggest", solve.Suggest)
	mux.HandleFunc("POST /api/rate", solve.Rate)
	return mux
}
