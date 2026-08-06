package solver

import (
	"errors"
	"fmt"
	"math"
	"runtime"
	"sort"
	"sync"

	"github.com/mmrzz/wordle-coach/pkg/data"
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

// Universe selects which words are allowed to be the answer.
type Universe int

const (
	// Official narrows within the curated answer list, which is what the
	// original game draws its solution from.
	Official Universe = iota
	// OffBook treats every legal guess as a possible answer. Some clones do
	// not honour the answer list, and against those the official list is not
	// merely unhelpful but wrong: it rules the solution out.
	OffBook
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
	// Universe is where the answer may come from. The zero value is the
	// official list, so going off the books is always a deliberate act.
	Universe Universe
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
	if o.Universe != Official && o.Universe != OffBook {
		return o, fmt.Errorf("solver: unknown universe %d", o.Universe)
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
	// Freq is the word's Zipf frequency, roughly 1 (rare) to 7 (common). It
	// never competes with the score, only breaks ties among words that tell
	// you exactly as much: those are not equally easy to think of.
	Freq float64
}

// Result is the answer to a suggest request.
type Result struct {
	// PossibleCount is the number of answers still consistent with history.
	PossibleCount int
	Suggestions   []Suggestion
	// Remaining lists the candidates, but only when there are few enough to
	// be worth showing, that is PossibleCount <= RemainingThreshold.
	Remaining []string
	// Letters is where each letter of the alphabet stands across the
	// candidates, for a player who would rather work it out than be told.
	Letters []LetterOdds
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

// ErrInconsistent reports a history not even a legal guess satisfies, which
// means the relayed feedback contradicts itself. Nearly always a mistyped
// colour, because no word anywhere could have produced it.
var ErrInconsistent = errors.New("solver: no word is consistent with this history")

// ErrOffBookOnly reports a history no official answer satisfies although some
// other legal guess does. It is the difference between "you coloured a tile
// wrong" and "this site is not using the official answer list", and only the
// player can say which, so it is reported rather than decided here.
var ErrOffBookOnly = errors.New("solver: no official answer is consistent with this history, but a legal guess is")

// Engine scores guesses against a corpus. It holds no per-game state, so a
// single Engine serves every request concurrently.
type Engine struct {
	set *data.Set
}

// New returns an Engine over set.
func New(set *data.Set) *Engine {
	return &Engine{set: set}
}

// answers returns every word u permits as a solution.
func (e *Engine) answers(u Universe) []string {
	if u == OffBook {
		return e.set.Allowed
	}
	return e.set.Answers
}

// Filter returns the words of u still consistent with history, in list order.
//
// A word survives exactly when replaying every past guess against it would
// have produced the feedback that was actually observed. Greens, yellows,
// greys and all the duplicate-letter arithmetic fall out of that one test, so
// nothing else needs tracking. Intersection commutes, so the order of history
// does not matter and the set is always rebuilt from the full answer list
// rather than carried between requests.
func (e *Engine) Filter(history []Turn, u Universe) []string {
	pool := e.answers(u)
	candidates := make([]string, 0, len(pool))
	for _, answer := range pool {
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
	// ranking is a constant and comes from a table: one per answer list, since
	// the two rank openers differently. A moved slider is a third ranking and
	// has to be computed.
	if len(history) == 0 && math.Abs(opts.Beta-DefaultBeta) < betaEpsilon {
		answers := e.answers(opts.Universe)
		return Result{
			PossibleCount: len(answers),
			Suggestions:   openers(opts.Limit, opts.Universe),
			// Cheap enough to compute even here: it is one pass over the
			// answers, against the pattern evaluations the table saves.
			Letters: letterOdds(answers),
		}, nil
	}

	candidates := e.Filter(history, opts.Universe)
	if len(candidates) == 0 {
		return Result{}, e.emptyErr(history, opts.Universe)
	}

	scored := e.scorePool(e.pool(history, opts.Mode), candidates, opts.Beta)
	sortSuggestions(scored)

	result := Result{
		PossibleCount: len(candidates),
		Suggestions:   scored[:min(opts.Limit, len(scored))],
		Letters:       letterOdds(candidates),
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

	candidates := e.Filter(history, opts.Universe)
	if len(candidates) == 0 {
		return Coach{}, e.emptyErr(history, opts.Universe)
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

// emptyErr says which kind of dead end an empty candidate set is.
//
// Off the books there is nothing wider left to try, so the feedback simply
// contradicts itself. On the official list the same emptiness has a second
// explanation — a game whose answer was never on that list — and the two ask
// different things of the player, so they get different errors.
func (e *Engine) emptyErr(history []Turn, u Universe) error {
	if u == Official && len(e.Filter(history, OffBook)) > 0 {
		return ErrOffBookOnly
	}
	return ErrInconsistent
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
					Freq:         e.set.Freq(guess),
				}
			}
		}(lo, hi)
	}
	wg.Wait()

	return out
}

// sortSuggestions ranks by information gained. Ties fall first to words that
// could win outright this turn, then to the more common word, then to the
// alphabet so the order is total and results are stable.
//
// Popularity earns its place there because two guesses worth the same number
// of bits are interchangeable to the engine and not at all to the player: a
// tie between SOARE and AROSE should not hand them the one they have to be
// told is a word.
func sortSuggestions(s []Suggestion) {
	sort.Slice(s, func(i, j int) bool {
		if s[i].Bits != s[j].Bits {
			return s[i].Bits > s[j].Bits
		}
		if s[i].InAnswerSet != s[j].InAnswerSet {
			return s[i].InAnswerSet
		}
		if s[i].Freq != s[j].Freq {
			return s[i].Freq > s[j].Freq
		}
		return s[i].Word < s[j].Word
	})
}
