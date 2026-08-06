package api

import (
	"errors"
	"net/http"

	"github.com/mmrzz/wordle-coach/internal/solver"
)

const (
	// defaultSuggestions is how many guesses a suggest request returns.
	defaultSuggestions = 5
	// maxSuggestions bounds the response so one request cannot ask for the
	// whole pool.
	maxSuggestions = 50
)

// Solver serves the suggest and rate endpoints.
//
// Requests carry the whole history and the server keeps nothing between them,
// so a restart cannot lose a game in progress and any instance can serve any
// request. Rebuilding the candidate set costs a few thousand pattern
// evaluations, which is nothing beside scoring the guess pool that follows.
type Solver struct {
	engine *solver.Engine
}

// NewSolver returns a handler set backed by engine.
func NewSolver(engine *solver.Engine) *Solver {
	return &Solver{engine: engine}
}

type suggestRequest struct {
	Mode    string        `json:"mode"`
	History []turnRequest `json:"history"`
	// Limit is how many suggestions to return, defaulting to 5.
	Limit int `json:"limit"`
	// Beta is the temperature slider. Absent means the default.
	Beta *float64 `json:"beta"`
}

type suggestResponse struct {
	// PossibleCount is how many answers remain, the headline number for the UI.
	PossibleCount int                  `json:"possibleCount"`
	Suggestions   []suggestionResponse `json:"suggestions"`
	// Remaining is sent only when few enough answers are left to list.
	Remaining []string `json:"remaining,omitempty"`
	// Letters is all 26 in order, so the UI can group and sort them however
	// it likes without having to know which ones were left out.
	Letters []letterResponse `json:"letters"`
}

// Suggest returns the best next guesses given the feedback so far.
func (s *Solver) Suggest(w http.ResponseWriter, r *http.Request) {
	var req suggestRequest
	if err := decodeBody(w, r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "bad_request", err)
		return
	}

	opts, err := decodeOptions(req.Mode, req.Beta, req.Limit)
	if err != nil {
		respondError(w, http.StatusBadRequest, "bad_options", err)
		return
	}
	history, err := decodeHistory(req.History)
	if err != nil {
		respondError(w, http.StatusBadRequest, "bad_history", err)
		return
	}

	result, err := s.engine.Suggest(history, opts)
	if err != nil {
		writeSolverError(w, err)
		return
	}

	respond(w, http.StatusOK, suggestResponse{
		PossibleCount: result.PossibleCount,
		Suggestions:   toSuggestions(result.Suggestions),
		Remaining:     result.Remaining,
		Letters:       toLetters(result.Letters),
	})
}

type rateRequest struct {
	Mode string `json:"mode"`
	// History holds the turns before Played, not including it: a word is
	// graded against what was known when it was played.
	History []turnRequest `json:"history"`
	Played  string        `json:"played"`
	Beta    *float64      `json:"beta"`
}

type rateResponse struct {
	Played             string             `json:"played"`
	PlayedBits         float64            `json:"playedBits"`
	PlayedExpRemaining float64            `json:"playedExpRemaining"`
	PlayedRank         int                `json:"playedRank"`
	Percentile         float64            `json:"percentile"`
	Best               suggestionResponse `json:"best"`
	GapBits            float64            `json:"gapBits"`
	PoolSize           int                `json:"poolSize"`
	PossibleCount      int                `json:"possibleCount"`
}

// Rate grades a played word against the position it was played from. The
// numbers go out raw; turning them into a grade and a sentence is the
// frontend's job, so the wording can change without touching Go.
func (s *Solver) Rate(w http.ResponseWriter, r *http.Request) {
	var req rateRequest
	if err := decodeBody(w, r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "bad_request", err)
		return
	}

	opts, err := decodeOptions(req.Mode, req.Beta, 1)
	if err != nil {
		respondError(w, http.StatusBadRequest, "bad_options", err)
		return
	}
	history, err := decodeHistory(req.History)
	if err != nil {
		respondError(w, http.StatusBadRequest, "bad_history", err)
		return
	}
	played, err := decodeWord(req.Played)
	if err != nil {
		respondError(w, http.StatusBadRequest, "bad_word", err)
		return
	}

	coach, err := s.engine.Rate(history, played, opts)
	if err != nil {
		writeSolverError(w, err)
		return
	}

	respond(w, http.StatusOK, rateResponse{
		Played:             coach.Played,
		PlayedBits:         round(coach.PlayedBits, 4),
		PlayedExpRemaining: round(coach.PlayedExpRemaining, 4),
		PlayedRank:         coach.PlayedRank,
		Percentile:         round(coach.Percentile, 2),
		Best:               toSuggestions([]solver.Suggestion{coach.Best})[0],
		GapBits:            round(coach.GapBits, 4),
		PoolSize:           coach.PoolSize,
		PossibleCount:      coach.PossibleCount,
	})
}

// writeSolverError maps engine failures onto status codes. An inconsistent
// history is a coherent request about an impossible position, which is what
// 422 is for, and it gets its own code because the UI should say "check the
// colours you entered" rather than showing a generic failure.
func writeSolverError(w http.ResponseWriter, err error) {
	if errors.Is(err, solver.ErrInconsistent) {
		respondError(w, http.StatusUnprocessableEntity, "inconsistent_history", err)
		return
	}
	respondError(w, http.StatusBadRequest, "bad_request", err)
}
