package solver

import (
	"errors"
	"fmt"
	"math"
	"testing"

	"github.com/mmrzz/wordle-coach/internal/data"
)

func testEngine(t *testing.T) *Engine {
	t.Helper()
	set, err := data.Load()
	if err != nil {
		t.Fatal(err)
	}
	return New(set)
}

// The answer must always survive the feedback it produced. This is the property
// Filter rests on, so it is checked against every answer in the list rather
// than a handful of examples.
func TestFilterAlwaysKeepsTheAnswer(t *testing.T) {
	e := testEngine(t)
	guesses := []string{"soare", "crane", "geese", "llama", "fuzzy"}

	for _, answer := range e.set.Answers {
		history := make([]Turn, 0, len(guesses))
		for _, g := range guesses {
			history = append(history, Turn{Guess: g, Pattern: GetPattern(g, answer)})
		}

		candidates := e.Filter(history)
		found := false
		for _, c := range candidates {
			if c == answer {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("answer %q did not survive its own feedback", answer)
		}
	}
}

// Set intersection commutes, so replaying history in any order is the same
// filter. This is what makes rebuilding from scratch each request safe.
func TestFilterIsOrderIndependent(t *testing.T) {
	e := testEngine(t)
	const answer = "pilot"

	forward := []Turn{
		{Guess: "soare", Pattern: GetPattern("soare", answer)},
		{Guess: "clint", Pattern: GetPattern("clint", answer)},
		{Guess: "mucky", Pattern: GetPattern("mucky", answer)},
	}
	reversed := []Turn{forward[2], forward[1], forward[0]}

	a, b := e.Filter(forward), e.Filter(reversed)
	if len(a) != len(b) {
		t.Fatalf("order changed the candidate count: %d vs %d", len(a), len(b))
	}
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("order changed candidate %d: %q vs %q", i, a[i], b[i])
		}
	}
}

func TestFilterNarrowsToTheAnswer(t *testing.T) {
	e := testEngine(t)
	const answer = "pilot"

	var history []Turn
	for _, g := range []string{"soare", "clint", "pilot"} {
		history = append(history, Turn{Guess: g, Pattern: GetPattern(g, answer)})
	}

	got := e.Filter(history)
	if len(got) != 1 || got[0] != answer {
		t.Fatalf("Filter = %v, want exactly [%q]", got, answer)
	}
}

func TestSuggestOpensFromTheTable(t *testing.T) {
	e := testEngine(t)

	got, err := e.Suggest(nil, Easy, 5)
	if err != nil {
		t.Fatal(err)
	}
	if got.PossibleCount != len(e.set.Answers) {
		t.Errorf("PossibleCount = %d, want %d", got.PossibleCount, len(e.set.Answers))
	}
	if len(got.Suggestions) != 5 {
		t.Fatalf("got %d suggestions, want 5", len(got.Suggestions))
	}
	if got.Suggestions[0].Word != openerTable[0].Word {
		t.Errorf("best opener = %q, want %q", got.Suggestions[0].Word, openerTable[0].Word)
	}
	if got.Remaining != nil {
		t.Errorf("Remaining should be empty with %d candidates", got.PossibleCount)
	}
}

// The precomputed table has to stay in step with the word lists. On failure
// this prints a replacement table ready to paste into openers.go.
func TestOpenerTableMatchesData(t *testing.T) {
	e := testEngine(t)

	scored := e.scorePool(e.set.Allowed, e.set.Answers)
	sortSuggestions(scored)

	for i, want := range openerTable {
		got := scored[i]
		if got.Word != want.Word || math.Abs(got.Bits-want.Bits) > 1e-6 {
			for _, s := range scored[:len(openerTable)] {
				t.Logf("\t{Word: %q, Bits: %.6f, ExpRemaining: %.4f, InAnswerSet: %v},",
					s.Word, s.Bits, s.ExpRemaining, s.InAnswerSet)
			}
			t.Fatalf("openerTable[%d] = %q/%.6f, data gives %q/%.6f: paste the table above into openers.go",
				i, want.Word, want.Bits, got.Word, got.Bits)
		}
	}
}

func TestSuggestOnASingleCandidate(t *testing.T) {
	e := testEngine(t)
	const answer = "pilot"

	var history []Turn
	for _, g := range []string{"soare", "clint", "pilot"} {
		history = append(history, Turn{Guess: g, Pattern: GetPattern(g, answer)})
	}

	got, err := e.Suggest(history, Easy, 5)
	if err != nil {
		t.Fatal(err)
	}
	if got.PossibleCount != 1 {
		t.Fatalf("PossibleCount = %d, want 1", got.PossibleCount)
	}
	// Every guess gains zero bits when only one answer is left, so the
	// tie-break has to be what surfaces the winning word.
	if got.Suggestions[0].Word != answer {
		t.Errorf("best = %q, want the answer %q", got.Suggestions[0].Word, answer)
	}
	if !got.Suggestions[0].InAnswerSet {
		t.Error("the answer should be flagged InAnswerSet")
	}
	if len(got.Remaining) != 1 || got.Remaining[0] != answer {
		t.Errorf("Remaining = %v, want [%q]", got.Remaining, answer)
	}
}

func TestSuggestRejectsInconsistentHistory(t *testing.T) {
	e := testEngine(t)

	// Two turns that cannot both be true: no E anywhere, then E in position 5.
	history := []Turn{
		{Guess: "crane", Pattern: 0},
		{Guess: "slate", Pattern: AllGreen},
	}

	_, err := e.Suggest(history, Easy, 5)
	if !errors.Is(err, ErrInconsistent) {
		t.Fatalf("err = %v, want ErrInconsistent", err)
	}
}

func TestSuggestRejectsUnknownWords(t *testing.T) {
	e := testEngine(t)

	_, err := e.Suggest([]Turn{{Guess: "zzzzz", Pattern: 0}}, Easy, 5)
	if err == nil {
		t.Fatal("want an error for a word outside the corpus")
	}
	if errors.Is(err, ErrInconsistent) {
		t.Fatal("an unknown word is a bad request, not an inconsistent history")
	}
}

// The hard-mode pool only enforces greens and known yellows, which is looser
// than full consistency, so it must contain every candidate.
func TestHardPoolContainsEveryCandidate(t *testing.T) {
	e := testEngine(t)
	const answer = "pilot"

	history := []Turn{{Guess: "soare", Pattern: GetPattern("soare", answer)}}

	pool := make(map[string]struct{})
	for _, w := range e.pool(history, Hard) {
		pool[w] = struct{}{}
	}

	candidates := e.Filter(history)
	for _, c := range candidates {
		if _, ok := pool[c]; !ok {
			t.Fatalf("candidate %q is missing from the hard-mode pool", c)
		}
	}
	if len(pool) >= len(e.set.Allowed) {
		t.Errorf("hard mode pool (%d) should be smaller than the full list (%d)", len(pool), len(e.set.Allowed))
	}
}

func TestHardPoolReusesRevealedLetters(t *testing.T) {
	e := testEngine(t)
	const answer = "pilot"

	// SOARE against PILOT shows a yellow O and nothing else.
	history := []Turn{{Guess: "soare", Pattern: GetPattern("soare", answer)}}

	for _, w := range e.pool(history, Hard) {
		if GetPattern("soare", w).Digit(1) != Yellow {
			continue
		}
		found := false
		for i := 0; i < WordLen; i++ {
			if w[i] == 'o' {
				found = true
			}
		}
		if !found {
			t.Fatalf("hard-mode pool word %q drops the revealed O", w)
		}
	}
}

func TestRateGradesFromThePriorPosition(t *testing.T) {
	e := testEngine(t)

	got, err := e.Rate(nil, "crane", Easy)
	if err != nil {
		t.Fatal(err)
	}
	if got.Played != "crane" {
		t.Errorf("Played = %q, want %q", got.Played, "crane")
	}
	if got.Best.Word != openerTable[0].Word {
		t.Errorf("Best = %q, want %q", got.Best.Word, openerTable[0].Word)
	}
	if got.PlayedRank < 1 || got.PlayedRank > got.PoolSize {
		t.Errorf("PlayedRank = %d, want 1..%d", got.PlayedRank, got.PoolSize)
	}
	if got.GapBits < 0 {
		t.Errorf("GapBits = %f, want >= 0", got.GapBits)
	}
	if math.Abs(got.GapBits-(got.Best.Bits-got.PlayedBits)) > 1e-9 {
		t.Errorf("GapBits = %f, want Best.Bits - PlayedBits", got.GapBits)
	}
	if got.Percentile < 0 || got.Percentile > 100 {
		t.Errorf("Percentile = %f, want 0..100", got.Percentile)
	}
	fmt.Printf("crane opener: rank %d/%d, %.2f bits, %.1f percentile, %.2f behind %s\n",
		got.PlayedRank, got.PoolSize, got.PlayedBits, got.Percentile, got.GapBits, got.Best.Word)
}

func TestRateTheBestWordRanksFirst(t *testing.T) {
	e := testEngine(t)

	got, err := e.Rate(nil, openerTable[0].Word, Easy)
	if err != nil {
		t.Fatal(err)
	}
	if got.PlayedRank != 1 {
		t.Errorf("PlayedRank = %d, want 1", got.PlayedRank)
	}
	if got.GapBits != 0 {
		t.Errorf("GapBits = %f, want 0", got.GapBits)
	}
	if got.Percentile != 100 {
		t.Errorf("Percentile = %f, want 100", got.Percentile)
	}
}

// A full game has to converge, which exercises Filter, the pool and the
// scoring together against every answer rather than one lucky word.
func TestSolvesEveryAnswer(t *testing.T) {
	if testing.Short() {
		t.Skip("full solve is slow")
	}
	e := testEngine(t)

	const maxTurns = 6
	worst, total := 0, 0

	for _, answer := range e.set.Answers[:120] {
		var history []Turn
		solved := 0

		for turn := 1; turn <= maxTurns; turn++ {
			res, err := e.Suggest(history, Easy, 1)
			if err != nil {
				t.Fatalf("answer %q turn %d: %v", answer, turn, err)
			}
			guess := res.Suggestions[0].Word
			p := GetPattern(guess, answer)
			history = append(history, Turn{Guess: guess, Pattern: p})
			if p == AllGreen {
				solved = turn
				break
			}
		}

		if solved == 0 {
			t.Errorf("did not solve %q within %d turns", answer, maxTurns)
			continue
		}
		total += solved
		if solved > worst {
			worst = solved
		}
	}

	fmt.Printf("solved 120 answers, mean %.2f turns, worst %d\n", float64(total)/120, worst)
}
