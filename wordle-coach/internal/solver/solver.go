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

// Suggest returns the n best guesses given the feedback so far. An empty
// history returns the precomputed openers, since with nothing yet known every
// game opens from the same position.
func (e *Engine) Suggest(history []Turn, mode Mode, n int) (Result, error) {
	if err := e.validate(history); err != nil {
		return Result{}, err
	}
	if n < 1 {
		n = 1
	}

	if len(history) == 0 {
		return Result{
			PossibleCount: len(e.set.Answers),
			Suggestions:   openers(n),
		}, nil
	}

	candidates := e.Filter(history)
	if len(candidates) == 0 {
		return Result{}, ErrInconsistent
	}

	scored := e.scorePool(e.pool(history, mode), candidates)
	sortSuggestions(scored)

	result := Result{
		PossibleCount: len(candidates),
		Suggestions:   scored[:min(n, len(scored))],
	}
	if len(candidates) <= RemainingThreshold {
		result.Remaining = candidates
	}
	return result, nil
}

// Rate grades played from the position it was actually played, so history must
// hold the turns before it and not the turn itself.
func (e *Engine) Rate(history []Turn, played string, mode Mode) (Coach, error) {
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

	pool := e.pool(history, mode)
	scored := e.scorePool(pool, candidates)
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
func (e *Engine) scorePool(pool, candidates []string) []Suggestion {
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

				// Both scores come off the same histogram in one pass, so the
				// engine stays neutral about which one the UI prefers.
				bits, expRemaining := 0.0, 0.0
				for _, n := range counts {
					if n == 0 {
						continue
					}
					p := float64(n) / total
					bits -= p * math.Log2(p)
					expRemaining += p * float64(n)
				}

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
