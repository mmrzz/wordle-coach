package solver

import (
	"math"
	"testing"
)

// histogram builds a counts array from bucket sizes, for testing collapse.
func histogram(sizes ...int) (*[NumPatterns]int, float64) {
	var counts [NumPatterns]int
	total := 0
	for i, n := range sizes {
		counts[i] = n
		total += n
	}
	return &counts, float64(total)
}

// At beta 2 the Renyi entropy is log2(N / expected remaining), so maximising
// the score is exactly minimising the expected number of answers left. This
// pins the slider's "safe" end to the metric the brief describes.
func TestCollapseAtBetaTwoTracksExpectedRemaining(t *testing.T) {
	for _, sizes := range [][]int{
		{10, 10, 10, 10},
		{37, 1, 1, 1},
		{100},
		{5, 3, 2, 1, 1, 1},
	} {
		counts, total := histogram(sizes...)
		bits, expRemaining := collapse(counts, total, 2)

		want := math.Log2(total / expRemaining)
		if math.Abs(bits-want) > 1e-9 {
			t.Errorf("sizes %v: bits = %f, want log2(N/expRemaining) = %f", sizes, bits, want)
		}
	}
}

// As beta approaches zero the score approaches log2 of the number of distinct
// patterns, ignoring how lopsided the split is.
func TestCollapseNearZeroCountsBuckets(t *testing.T) {
	counts, total := histogram(97, 1, 1, 1)
	bits, _ := collapse(counts, total, MinBeta)

	want := math.Log2(4)
	if math.Abs(bits-want) > 0.05 {
		t.Errorf("bits at beta=%g = %f, want about log2(4) = %f", MinBeta, bits, want)
	}
}

// As beta grows the score approaches minimax, -log2 of the largest share. The
// approach is gradual, so the test is that raising beta moves towards that
// limit rather than reaching it: at MaxBeta the score is far nearer minimax
// than Shannon entropy is.
func TestCollapseAtHighBetaApproachesMinimax(t *testing.T) {
	counts, total := histogram(50, 30, 20)
	minimax := -math.Log2(0.5)

	shannon, _ := collapse(counts, total, DefaultBeta)
	high, _ := collapse(counts, total, MaxBeta)

	if math.Abs(high-minimax) >= math.Abs(shannon-minimax) {
		t.Errorf("beta %g (%f) is no closer to minimax %f than beta 1 (%f)",
			MaxBeta, high, minimax, shannon)
	}
	if math.Abs(high-minimax) > 0.15 {
		t.Errorf("bits at beta=%g = %f, want within 0.15 of %f", MaxBeta, high, minimax)
	}
}

// A uniform split is the same score at every beta: all the Renyi entropies
// agree when there is nothing lopsided to disagree about.
func TestCollapseAgreesOnUniformSplits(t *testing.T) {
	counts, total := histogram(25, 25, 25, 25)
	want := 2.0

	for _, beta := range []float64{MinBeta, 0.5, DefaultBeta, 2, 4, MaxBeta} {
		bits, _ := collapse(counts, total, beta)
		if math.Abs(bits-want) > 1e-9 {
			t.Errorf("beta %g: bits = %f, want %f", beta, bits, want)
		}
	}
}

// Renyi entropy never increases with beta, which is what makes the slider a
// single ordered axis from bold to safe rather than an arbitrary dial.
func TestCollapseIsMonotoneInBeta(t *testing.T) {
	counts, total := histogram(60, 20, 10, 5, 3, 2)

	previous := math.Inf(1)
	for _, beta := range []float64{MinBeta, 0.25, 0.5, DefaultBeta, 1.5, 2, 3, 5, MaxBeta} {
		bits, _ := collapse(counts, total, beta)
		if bits > previous+1e-9 {
			t.Errorf("beta %g: bits = %f rose above the previous %f", beta, bits, previous)
		}
		previous = bits
	}
}

// Expected remaining describes the position, not the objective, so it must not
// move when the slider does.
func TestCollapseExpectedRemainingIgnoresBeta(t *testing.T) {
	counts, total := histogram(60, 20, 10, 5, 3, 2)

	_, want := collapse(counts, total, DefaultBeta)
	for _, beta := range []float64{MinBeta, 0.5, 2, MaxBeta} {
		if _, got := collapse(counts, total, beta); math.Abs(got-want) > 1e-9 {
			t.Errorf("beta %g: expRemaining = %f, want %f", beta, got, want)
		}
	}
}

func TestOptionsRejectUnsetBeta(t *testing.T) {
	// The zero value must fail rather than quietly score at beta 0.
	if _, err := (Options{Mode: Easy}).check(); err == nil {
		t.Error("Options with no beta should be rejected")
	}
	for _, beta := range []float64{-1, MaxBeta + 1} {
		if _, err := (Options{Mode: Easy, Beta: beta}).check(); err == nil {
			t.Errorf("beta %g should be rejected", beta)
		}
	}
	if _, err := (Options{Mode: Hard, Beta: DefaultBeta}).check(); err != nil {
		t.Errorf("valid options rejected: %v", err)
	}
}

// The slider has to actually change the advice, or it is decoration. Turn one
// is the position to check it from: it is the same for every game, so this
// cannot pass or fail by luck of the chosen history.
//
// Not every position responds to the slider. Where one word dominates on every
// measure it stays top throughout, which is correct rather than a fault.
func TestBetaChangesTheSuggestion(t *testing.T) {
	e := testEngine(t)

	seen := make(map[string]bool)
	for _, beta := range []float64{MinBeta, 0.5, DefaultBeta, 2, 5, MaxBeta} {
		got, err := e.Suggest(nil, Options{Mode: Easy, Beta: beta, Limit: 1})
		if err != nil {
			t.Fatalf("beta %g: %v", beta, err)
		}
		best := got.Suggestions[0]
		seen[best.Word] = true
		t.Logf("beta %-5g -> %s (%.3f bits, %.1f expected left)",
			beta, best.Word, best.Bits, best.ExpRemaining)
	}

	if len(seen) < 2 {
		t.Error("every beta opened with the same word, so the slider does nothing")
	}
}

// Beta 2 is the expected-remaining objective, so its top word must be the one
// that genuinely minimises expected remaining across the whole pool. This is
// the strongest check that the slider's "safe" end means what it claims.
func TestBetaTwoMinimisesExpectedRemaining(t *testing.T) {
	e := testEngine(t)

	scored := e.scorePool(e.set.Allowed, e.set.Answers, 2)
	sortSuggestions(scored)
	best := scored[0]

	for _, s := range scored {
		if s.ExpRemaining < best.ExpRemaining-1e-9 {
			t.Fatalf("beta 2 chose %q (%.4f expected left) but %q leaves %.4f",
				best.Word, best.ExpRemaining, s.Word, s.ExpRemaining)
		}
	}
	t.Logf("beta 2 opener: %s, leaving %.2f of %d answers", best.Word, best.ExpRemaining, len(e.set.Answers))
}

// Turn one is served from the table only at the default beta; a moved slider
// has to be computed or it would silently return the wrong ranking.
func TestOpenersOnlyServeTheDefaultBeta(t *testing.T) {
	e := testEngine(t)

	fromTable, err := e.Suggest(nil, easy(1))
	if err != nil {
		t.Fatal(err)
	}
	if fromTable.Suggestions[0].Word != openerTable[0].Word {
		t.Fatalf("default beta = %q, want the table's %q", fromTable.Suggestions[0].Word, openerTable[0].Word)
	}

	computed, err := e.Suggest(nil, Options{Mode: Easy, Beta: 4, Limit: 1})
	if err != nil {
		t.Fatal(err)
	}
	if computed.Suggestions[0].Bits == fromTable.Suggestions[0].Bits {
		t.Error("beta 4 returned the table's score, so the table was served for the wrong beta")
	}
	t.Logf("opener at beta 4: %s (%.3f bits)", computed.Suggestions[0].Word, computed.Suggestions[0].Bits)
}
