package api

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"strings"

	"github.com/mmrzz/wordle-coach/internal/solver"
)

// maxHistory caps a request at more turns than any real game, so a malicious
// body cannot make the server score an unbounded number of positions.
const maxHistory = 12

// maxBody bounds the request body. A full history is well under a kilobyte.
const maxBody = 8 << 10

// turnRequest is one played guess and the colours the game showed for it.
type turnRequest struct {
	Guess string `json:"guess"`
	// Pattern is five characters of g (green), y (yellow) or b (black/grey),
	// left to right.
	Pattern string `json:"pattern"`
}

// errorResponse carries a machine-readable code so the UI can tell a mistyped
// colour from a mistyped word.
type errorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

// suggestionResponse mirrors solver.Suggestion. The engine returns raw numbers
// and the wording of any grade is left to the frontend.
type suggestionResponse struct {
	Word         string  `json:"word"`
	Bits         float64 `json:"bits"`
	ExpRemaining float64 `json:"expRemaining"`
	InAnswerSet  bool    `json:"inAnswerSet"`
}

// letterResponse mirrors solver.LetterOdds. Odds travel as shares of one
// rather than percentages, leaving the wording to the frontend.
type letterResponse struct {
	Letter   string  `json:"letter"`
	Presence float64 `json:"presence"`
	// Position counts from 0, and is -1 when the letter is ruled out.
	Position     int     `json:"position"`
	PositionOdds float64 `json:"positionOdds"`
}

// decodePattern converts the wire spelling into the engine's base-3 code, so
// the solver never sees a string. Case is ignored; nothing else is.
func decodePattern(s string) (solver.Pattern, error) {
	if len(s) != solver.WordLen {
		return 0, fmt.Errorf("pattern %q is %d characters, want %d", s, len(s), solver.WordLen)
	}

	code := solver.Pattern(0)
	for i := solver.WordLen - 1; i >= 0; i-- {
		var digit solver.Pattern
		switch s[i] {
		case 'g', 'G':
			digit = solver.Green
		case 'y', 'Y':
			digit = solver.Yellow
		case 'b', 'B':
			digit = solver.Grey
		default:
			return 0, fmt.Errorf("pattern %q contains %q, want only g, y or b", s, s[i])
		}
		code = code*3 + digit
	}
	return code, nil
}

// decodeHistory converts the wire history into engine turns. Words are lowered
// here so the frontend may send either case.
func decodeHistory(history []turnRequest) ([]solver.Turn, error) {
	if len(history) > maxHistory {
		return nil, fmt.Errorf("history has %d turns, the limit is %d", len(history), maxHistory)
	}

	turns := make([]solver.Turn, 0, len(history))
	for i, h := range history {
		pattern, err := decodePattern(h.Pattern)
		if err != nil {
			return nil, fmt.Errorf("history[%d]: %w", i, err)
		}
		turns = append(turns, solver.Turn{
			Guess:   strings.ToLower(strings.TrimSpace(h.Guess)),
			Pattern: pattern,
		})
	}
	return turns, nil
}

// decodeWord normalises a single word from the wire. The engine checks it
// against the corpus; this only rejects what would be a malformed request.
func decodeWord(s string) (string, error) {
	word := strings.ToLower(strings.TrimSpace(s))
	if len(word) != solver.WordLen {
		return "", fmt.Errorf("word %q is %d letters, want %d", s, len(word), solver.WordLen)
	}
	return word, nil
}

// decodeMode reads the guessing mode, defaulting to easy when unset.
func decodeMode(s string) (solver.Mode, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "easy":
		return solver.Easy, nil
	case "hard":
		return solver.Hard, nil
	default:
		return 0, fmt.Errorf("mode %q is not easy or hard", s)
	}
}

// decodeOptions builds the engine request. Beta arrives as a pointer so that an
// absent field means "default" while an explicit 0 is still rejected as out of
// range, rather than the two being indistinguishable.
func decodeOptions(mode string, beta *float64, limit int) (solver.Options, error) {
	m, err := decodeMode(mode)
	if err != nil {
		return solver.Options{}, err
	}

	b := solver.DefaultBeta
	if beta != nil {
		b = *beta
	}
	if b < solver.MinBeta || b > solver.MaxBeta {
		return solver.Options{}, fmt.Errorf("beta %g is outside %g..%g", b, solver.MinBeta, solver.MaxBeta)
	}

	if limit < 1 {
		limit = defaultSuggestions
	}
	if limit > maxSuggestions {
		return solver.Options{}, fmt.Errorf("limit %d is above the maximum of %d", limit, maxSuggestions)
	}

	return solver.Options{Mode: m, Beta: b, Limit: limit}, nil
}

// toSuggestions converts engine results for the wire, rounding the scores to a
// sane number of digits so the payload does not carry float noise.
func toSuggestions(in []solver.Suggestion) []suggestionResponse {
	out := make([]suggestionResponse, len(in))
	for i, s := range in {
		out[i] = suggestionResponse{
			Word:         s.Word,
			Bits:         round(s.Bits, 4),
			ExpRemaining: round(s.ExpRemaining, 4),
			InAnswerSet:  s.InAnswerSet,
		}
	}
	return out
}

// toLetters converts the per-letter odds for the wire.
func toLetters(in []solver.LetterOdds) []letterResponse {
	out := make([]letterResponse, len(in))
	for i, l := range in {
		out[i] = letterResponse{
			Letter:       l.Letter,
			Presence:     round(l.Presence, 4),
			Position:     l.Position,
			PositionOdds: round(l.PositionOdds, 4),
		}
	}
	return out
}

func round(f float64, places int) float64 {
	scale := math.Pow10(places)
	return math.Round(f*scale) / scale
}

// decodeBody reads a JSON request body of at most maxBody bytes, rejecting
// unknown fields so a typo in the frontend fails loudly instead of silently.
func decodeBody(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBody)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return fmt.Errorf("invalid request body: %w", err)
	}
	return nil
}

// respond writes v as JSON with the given status.
func respond(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	// The status is already sent, so a failure here can only be logged.
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("api: encoding response: %v", err)
	}
}

// respondError writes a failure with a code the UI can branch on.
func respondError(w http.ResponseWriter, status int, code string, err error) {
	respond(w, status, errorResponse{Error: err.Error(), Code: code})
}
