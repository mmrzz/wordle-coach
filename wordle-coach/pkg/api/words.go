package api

import (
	"net/http"

	"github.com/mmrzz/wordle-coach/pkg/data"
)

// Words serves the guess list so the browser can reject a typed non-word
// without a round trip, which is what makes the shake feel instant. The list is
// baked into the binary and never changes, so it is cached hard.
type Words struct {
	set *data.Set
}

// NewWords returns a handler serving set's allowed list.
func NewWords(set *data.Set) *Words {
	return &Words{set: set}
}

type wordsResponse struct {
	Allowed []string `json:"allowed"`
	// AnswerCount is how many of them can be the solution, for the UI's
	// opening "N possible answers" readout.
	AnswerCount int `json:"answerCount"`
}

// List returns every allowed guess.
func (h *Words) List(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "public, max-age=86400, immutable")

	respond(w, http.StatusOK, wordsResponse{
		Allowed:     h.set.Allowed,
		AnswerCount: len(h.set.Answers),
	})
}
