package solver

import (
	"errors"
	"fmt"
	"math"
	"runtime"
	"sort"
	"sync"

	"github.com/mmrzz/wordle-coach/internal/data"
)

// RemainingThreshold is the largest candidate count worth listing outright. At
// or below it the UI can simply show the words instead of a suggestion.
const RemainingThreshold = 10

// Mode selects which words may be guessed.
type Mode int

const (
	// Easy allows any word in the allowed list.
	Easy Mode = iota
	// Hard allows only words that reuse revealed greens and yellows. Note
	// this is a looser test than full consistency, so the hard-mode pool is a
	// superset of the candidate set.
	Hard
)

// Turn is one played guess and the feedback that came back from the game.
type Turn struct {
	Guess   string
	Pattern Pattern
}

// Options configures one request. There is no useful zero value: Beta must be
// set, so that forgetting it is an error rather than a silent rescoring.
type Options struct {
	Mode Mode
	// Beta is the Rényi order, between MinBeta and MaxBeta. See score.go.
	Beta float64
	// Limit is how many suggestions to return. Ignored by Rate.
	Limit int
}

// check validates the options and applies the one safe default.
func (o Options) check() (Options, error) {
	if o.Beta < MinBeta || o.Beta > MaxBeta {
		return o, fmt.Errorf("solver: beta %g is outside %g..%g", o.Beta, MinBeta, MaxBeta)
	}
	if o.Mode != Easy && o.Mode != Hard {
		return o, fmt.Errorf("solver: unknown mode %d", o.Mode)
	}
	if o.Limit < 1 {
		o.Limit = 1
	}
	return o, nil
}

// Suggestion is one scored guess.
type Suggestion struct {
	Word string
	// Bits is the expected information gained, higher is better.
	Bits float64
	// ExpRemaining is the expected size of the candidate set afterwards,
	// lower is better.
	ExpRemaining float64
	// InAnswerSet reports whether the word is itself still a candidate, so it
	// could win outright this turn.
	InAnswerSet bool
}

// Result is the answer to a suggest request.
type Result struct {
	// PossibleCount is the number of answers still consistent with history.
	PossibleCount int
	Suggestions   []Suggestion
	// Remaining lists the candidates, but only when there are few enough to
	// be worth showing, that is PossibleCount <= RemainingThreshold.
	Remaining []string
}

// Coach grades one played word against the position it was played from.
type Coach struct {
	Played             string
	PlayedBits         float64
	PlayedExpRemaining float64
	// PlayedRank is 1 for the best guess in the pool. Words scoring equally
	// share a rank.
	PlayedRank int
	// Percentile is the share of the pool the played word beat, 0 to 100.
	Percentile float64
	Best       Suggestion
	// GapBits is how many bits the best guess would have gained over the one
	// played, never negative.
	GapBits       float64
	PoolSize      int
	PossibleCount int
}

// ErrInconsistent reports a history no answer satisfies, which means the
// relayed feedback contradicts itself. Nearly always a mistyped colour.
var ErrInconsistent = errors.New("solver: no answer is consistent with this history")

// Engine scores guesses against a corpus. It holds no per-game state, so a
// single Engine serves every request concurrently.
type Engine struct {
	set *data.Set
}

// New returns an Engine over set.
func New(set *data.Set) *Engine {
	return &Engine{set: set}
}

// Filter returns the answers still consistent with history, in list order.
//
// A word survives exactly when replaying every past guess against it would
// have produced the feedback that was actually observed. Greens, yellows,
// greys and all the duplicate-letter arithmetic fall out of that one test, so
// nothing else needs tracking. Intersection commutes, so the order of history
// does not matter and the set is always rebuilt from the full answer list
// rather than carried between requests.
func (e *Engine) Filter(history []Turn) []string {
	candidates := make([]string, 0, len(e.set.Answers))
	for _, answer := range e.set.Answers {
		consistent := true
		for _, turn := range history {
			if GetPattern(turn.Guess, answer) != turn.Pattern {
				consistent = false
				break
			}
		}
		if consistent {
			candidates = append(candidates, answer)
		}
	}
	return candidates
}

// Suggest returns the best guesses given the feedback so far.
func (e *Engine) Suggest(history []Turn, opts Options) (Result, error) {
	opts, err := opts.check()
	if err != nil {
		return Result{}, err
	}
	if err := e.validate(history); err != nil {
		return Result{}, err
	}

	// With nothing known every game opens from the same position, so the
	// ranking is a constant and comes from the table. A moved slider is a
	// different ranking, which has to be computed.
	if len(history) == 0 && math.Abs(opts.Beta-DefaultBeta) < betaEpsilon {
		return Result{
			PossibleCount: len(e.set.Answers),
			Suggestions:   openers(opts.Limit),
		}, nil
	}

	candidates := e.Filter(history)
	if len(candidates) == 0 {
		return Result{}, ErrInconsistent
	}

	scored := e.scorePool(e.pool(history, opts.Mode), candidates, opts.Beta)
	sortSuggestions(scored)

	result := Result{
		PossibleCount: len(candidates),
		Suggestions:   scored[:min(opts.Limit, len(scored))],
	}
	if len(candidates) <= RemainingThreshold {
		result.Remaining = candidates
	}
	return result, nil
}

// Rate grades played from the position it was actually played, so history must
// hold the turns before it and not the turn itself.
func (e *Engine) Rate(history []Turn, played string, opts Options) (Coach, error) {
	opts, err := opts.check()
	if err != nil {
		return Coach{}, err
	}
	if err := e.validate(history); err != nil {
		return Coach{}, err
	}
	if !e.set.IsAllowed(played) {
		return Coach{}, fmt.Errorf("solver: %q is not an allowed guess", played)
	}

	candidates := e.Filter(history)
	if len(candidates) == 0 {
		return Coach{}, ErrInconsistent
	}

	scored := e.scorePool(e.pool(history, opts.Mode), candidates, opts.Beta)
	sortSuggestions(scored)

	var play Suggestion
	found := false
	for _, s := range scored {
		if s.Word == played {
			play, found = s, true
			break
		}
	}
	if !found {
		return Coach{}, fmt.Errorf("solver: %q was not legal in hard mode at this point", played)
	}

	// Equal scores share a rank, so a word tied for best ranks 1st rather
	// than being penalised for its position in the sort.
	rank := 1
	for _, s := range scored {
		if s.Bits > play.Bits {
			rank++
		}
	}

	percentile := 100.0
	if len(scored) > 1 {
		percentile = 100 * float64(len(scored)-rank) / float64(len(scored)-1)
	}

	return Coach{
		Played:             played,
		PlayedBits:         play.Bits,
		PlayedExpRemaining: play.ExpRemaining,
		PlayedRank:         rank,
		Percentile:         percentile,
		Best:               scored[0],
		GapBits:            math.Max(0, scored[0].Bits-play.Bits),
		PoolSize:           len(scored),
		PossibleCount:      len(candidates),
	}, nil
}

// validate rejects guesses the corpus does not contain, which would otherwise
// index out of range in GetPattern.
func (e *Engine) validate(history []Turn) error {
	for i, turn := range history {
		if !e.set.IsAllowed(turn.Guess) {
			return fmt.Errorf("solver: history[%d]: %q is not an allowed guess", i, turn.Guess)
		}
		if turn.Pattern > AllGreen {
			return fmt.Errorf("solver: history[%d]: pattern %d is out of range", i, turn.Pattern)
		}
	}
	return nil
}

// pool returns the words that may be guessed next.
func (e *Engine) pool(history []Turn, mode Mode) []string {
	if mode != Hard || len(history) == 0 {
		return e.set.Allowed
	}

	// Hard mode must reuse what has been revealed: greens stay put, and every
	// letter shown green or yellow reappears at least as many times as it was
	// shown in any single turn.
	var greens [WordLen]byte
	var required [26]int
	for _, turn := range history {
		var seen [26]int
		for i := 0; i < WordLen; i++ {
			switch turn.Pattern.Digit(i) {
			case Green:
				greens[i] = turn.Guess[i]
				seen[turn.Guess[i]-'a']++
			case Yellow:
				seen[turn.Guess[i]-'a']++
			}
		}
		for c, n := range seen {
			if n > required[c] {
				required[c] = n
			}
		}
	}

	pool := make([]string, 0, len(e.set.Allowed))
word:
	for _, w := range e.set.Allowed {
		var have [26]int
		for i := 0; i < WordLen; i++ {
			if greens[i] != 0 && w[i] != greens[i] {
				continue word
			}
			have[w[i]-'a']++
		}
		for c, n := range required {
			if have[c] < n {
				continue word
			}
		}
		pool = append(pool, w)
	}
	return pool
}

// scorePool buckets candidates by the pattern each pool word would produce and
// collapses each histogram into both scores. The work splits cleanly across
// cores because every pool word is independent.
func (e *Engine) scorePool(pool, candidates []string, beta float64) []Suggestion {
	inCandidates := make(map[string]struct{}, len(candidates))
	for _, w := range candidates {
		inCandidates[w] = struct{}{}
	}

	out := make([]Suggestion, len(pool))
	total := float64(len(candidates))

	workers := min(runtime.NumCPU(), len(pool))
	chunk := (len(pool) + workers - 1) / workers

	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		lo := w * chunk
		hi := min(lo+chunk, len(pool))
		if lo >= hi {
			break
		}
		wg.Add(1)
		go func(lo, hi int) {
			defer wg.Done()
			for i := lo; i < hi; i++ {
				guess := pool[i]

				var counts [NumPatterns]int
				for _, answer := range candidates {
					counts[GetPattern(guess, answer)]++
				}

				bits, expRemaining := collapse(&counts, total, beta)

				_, isCandidate := inCandidates[guess]
				out[i] = Suggestion{
					Word:         guess,
					Bits:         bits,
					ExpRemaining: expRemaining,
					InAnswerSet:  isCandidate,
				}
			}
		}(lo, hi)
	}
	wg.Wait()

	return out
}

// sortSuggestions ranks by information gained, breaking ties towards words
// that could win this turn and then alphabetically, so results are stable.
func sortSuggestions(s []Suggestion) {
	sort.Slice(s, func(i, j int) bool {
		if s[i].Bits != s[j].Bits {
			return s[i].Bits > s[j].Bits
		}
		if s[i].InAnswerSet != s[j].InAnswerSet {
			return s[i].InAnswerSet
		}
		return s[i].Word < s[j].Word
	})
}
