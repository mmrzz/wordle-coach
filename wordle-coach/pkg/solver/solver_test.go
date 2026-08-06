package solver

import (
	"errors"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/mmrzz/wordle-coach/pkg/data"
)

func testEngine(t *testing.T) *Engine {
	t.Helper()
	set, err := data.Load()
	if err != nil {
		t.Fatal(err)
	}
	return New(set)
}

// easy is the default request: easy mode, Shannon entropy, n suggestions.
func easy(n int) Options {
	return Options{Mode: Easy, Beta: DefaultBeta, Limit: n}
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

		candidates := e.Filter(history, Official)
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

	a, b := e.Filter(forward, Official), e.Filter(reversed, Official)
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

	got := e.Filter(history, Official)
	if len(got) != 1 || got[0] != answer {
		t.Fatalf("Filter = %v, want exactly [%q]", got, answer)
	}
}

func TestSuggestOpensFromTheTable(t *testing.T) {
	e := testEngine(t)

	got, err := e.Suggest(nil, easy(5))
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

// The two rankings are constants, so they are tabulated rather than found
// again on every new game. This is what keeps them honest: each table must be
// what scoring the whole pool against its own answer list actually gives.
func TestOpenerTableMatchesData(t *testing.T) {
	e := testEngine(t)

	tables := []struct {
		name  string
		table []Suggestion
		u     Universe
	}{
		{"openerTable", openerTable, Official},
		{"offBookOpenerTable", offBookOpenerTable, OffBook},
	}

	for _, tt := range tables {
		scored := e.scorePool(e.set.Allowed, e.answers(tt.u), DefaultBeta)
		sortSuggestions(scored)

		for i, want := range tt.table {
			got := scored[i]
			if got.Word != want.Word || math.Abs(got.Bits-want.Bits) > 1e-6 {
				for _, s := range scored[:len(tt.table)] {
					t.Logf("	{Word: %q, Bits: %.6f, ExpRemaining: %.4f, InAnswerSet: %v},",
						s.Word, s.Bits, s.ExpRemaining, s.InAnswerSet)
				}
				t.Fatalf("%s[%d] = %q/%.6f, data gives %q/%.6f: paste the table above into openers.go",
					tt.name, i, want.Word, want.Bits, got.Word, got.Bits)
			}
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

	got, err := e.Suggest(history, easy(5))
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

	_, err := e.Suggest(history, easy(5))
	if !errors.Is(err, ErrInconsistent) {
		t.Fatalf("err = %v, want ErrInconsistent", err)
	}
}

func TestSuggestRejectsUnknownWords(t *testing.T) {
	e := testEngine(t)

	_, err := e.Suggest([]Turn{{Guess: "zzzzz", Pattern: 0}}, easy(5))
	if err == nil {
		t.Fatal("want an error for a word outside the corpus")
	}
	if errors.Is(err, ErrInconsistent) {
		t.Fatal("an unknown word is a bad request, not an inconsistent history")
	}
}

// Words that tell you the same amount are ordered by how common they are, so
// the one the player has actually heard of comes first.
func TestTiesBreakTowardsTheCommonerWord(t *testing.T) {
	// Equal bits and both outside the answer set, so popularity is the only
	// thing left to separate them.
	scored := []Suggestion{
		{Word: "aalii", Bits: 4.2, Freq: 1.1},
		{Word: "later", Bits: 4.2, Freq: 5.3},
		{Word: "zizel", Bits: 4.2, Freq: 1.0},
	}
	sortSuggestions(scored)

	if scored[0].Word != "later" {
		t.Errorf("best = %q, want %q: it is by far the commonest of the three", scored[0].Word, "later")
	}
	for i := 1; i < len(scored); i++ {
		if scored[i-1].Freq < scored[i].Freq {
			t.Errorf("order %v puts a rarer word ahead of a commoner one", words(scored))
			break
		}
	}
}

// Popularity only ever breaks a tie: it must never lift a worse guess above a
// better one, however famous the word.
func TestPopularityNeverOutranksInformation(t *testing.T) {
	scored := []Suggestion{
		{Word: "about", Bits: 3.0, Freq: 6.0},
		{Word: "aalii", Bits: 4.0, Freq: 1.0},
	}
	sortSuggestions(scored)

	if scored[0].Word != "aalii" {
		t.Errorf("best = %q, want the higher-scoring %q", scored[0].Word, "aalii")
	}
}

// The tie-break is only worth anything if the engine actually knows how common
// its words are.
func TestScorePoolCarriesFrequency(t *testing.T) {
	e := testEngine(t)

	scored := e.scorePool([]string{"about", "aalii"}, e.set.Answers, DefaultBeta)
	if scored[0].Freq <= scored[1].Freq {
		t.Errorf("ABOUT (%g) should read as commoner than AALII (%g)", scored[0].Freq, scored[1].Freq)
	}
}

func words(s []Suggestion) []string {
	out := make([]string, len(s))
	for i, x := range s {
		out[i] = x.Word
	}
	return out
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

	candidates := e.Filter(history, Official)
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

	got, err := e.Rate(nil, "crane", easy(1))
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

	got, err := e.Rate(nil, openerTable[0].Word, easy(1))
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
			res, err := e.Suggest(history, easy(1))
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

// offBookWord is a legal guess the official answer list does not contain, so
// filtering within that list can never find it. It is what a game not using
// the list can still be hiding.
const offBookWord = "aalii"

func TestOffBookAnswerIsFoundOutsideTheList(t *testing.T) {
	e := testEngine(t)
	if !e.set.IsAllowed(offBookWord) {
		t.Fatalf("%q is not a legal guess, so the fixture is wrong", offBookWord)
	}

	// The word played against itself: every tile green, which no other word in
	// the corpus can satisfy.
	history := []Turn{{Guess: offBookWord, Pattern: AllGreen}}

	if got := e.Filter(history, Official); len(got) != 0 {
		t.Fatalf("Filter(Official) = %v, want nothing: the answer is not on that list", got)
	}
	got := e.Filter(history, OffBook)
	if len(got) != 1 || got[0] != offBookWord {
		t.Fatalf("Filter(OffBook) = %v, want [%q]", got, offBookWord)
	}
}

// The whole point of the offer: the position the official list calls
// impossible has to become solvable once the books are set aside.
func TestSuggestOffTheBooksSolvesWhatTheListCannot(t *testing.T) {
	e := testEngine(t)
	history := []Turn{{Guess: offBookWord, Pattern: AllGreen}}

	opts := easy(5)
	opts.Universe = OffBook
	got, err := e.Suggest(history, opts)
	if err != nil {
		t.Fatalf("Suggest off the books: %v", err)
	}
	if got.PossibleCount != 1 {
		t.Errorf("PossibleCount = %d, want 1", got.PossibleCount)
	}
	if len(got.Remaining) != 1 || got.Remaining[0] != offBookWord {
		t.Errorf("Remaining = %v, want [%q]", got.Remaining, offBookWord)
	}
}

// A position only the wider list can explain must not be reported as a
// mistyped colour: the player is about to be asked which of the two it is, and
// the question only makes sense if the engine can tell them apart.
func TestOfficialUniverseSaysWhenOffBookWouldHelp(t *testing.T) {
	e := testEngine(t)
	history := []Turn{{Guess: offBookWord, Pattern: AllGreen}}

	_, err := e.Suggest(history, easy(5))
	if !errors.Is(err, ErrOffBookOnly) {
		t.Fatalf("Suggest err = %v, want ErrOffBookOnly", err)
	}

	_, err = e.Rate(history, "crane", easy(1))
	if !errors.Is(err, ErrOffBookOnly) {
		t.Fatalf("Rate err = %v, want ErrOffBookOnly", err)
	}
}

// Feedback that contradicts itself is impossible for every word there is, so
// widening the list changes nothing and the offer must not be made.
func TestContradictoryColoursStayInconsistentOffTheBooks(t *testing.T) {
	e := testEngine(t)

	// No E anywhere, then E in position 5.
	history := []Turn{
		{Guess: "crane", Pattern: 0},
		{Guess: "slate", Pattern: AllGreen},
	}

	for _, u := range []Universe{Official, OffBook} {
		opts := easy(5)
		opts.Universe = u
		_, err := e.Suggest(history, opts)
		if !errors.Is(err, ErrInconsistent) {
			t.Errorf("universe %d: err = %v, want ErrInconsistent", u, err)
		}
		if errors.Is(err, ErrOffBookOnly) {
			t.Errorf("universe %d: offered to go off the books, but nothing fits there either", u)
		}
	}
}

// An empty board off the books is the most expensive question the engine can
// be asked and the least informative, so it comes off a table exactly as the
// official opening does — and off its own table, not the official one.
func TestOffBookOpensFromItsOwnTable(t *testing.T) {
	e := testEngine(t)

	opts := easy(5)
	opts.Universe = OffBook

	start := time.Now()
	got, err := e.Suggest(nil, opts)
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("Suggest: %v", err)
	}

	if got.PossibleCount != len(e.set.Allowed) {
		t.Errorf("PossibleCount = %d, want every legal guess (%d)", got.PossibleCount, len(e.set.Allowed))
	}
	if got.Suggestions[0].Word == openerTable[0].Word {
		t.Errorf("best = %q, which is the official opener: the wider list ranks openers differently",
			got.Suggestions[0].Word)
	}
	// Scoring it live takes seconds even across every core, so the clock is
	// what proves a table answered instead.
	if elapsed > time.Second {
		t.Errorf("took %v, so the ranking was computed rather than tabulated", elapsed)
	}
}
