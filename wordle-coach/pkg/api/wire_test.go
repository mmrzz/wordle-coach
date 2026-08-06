package api

import (
	"testing"

	"github.com/mmrzz/wordle-coach/pkg/solver"
)

// The wire spelling and the engine's encoding are written in two places, so
// this pins them together: a hand-written pattern must equal the one the engine
// derives for the same guess and answer. If either side flips its digit order,
// this fails instead of silently corrupting every filter.
func TestDecodePatternMatchesEngine(t *testing.T) {
	tests := []struct {
		guess, answer, pattern string
	}{
		{"crane", "crane", "ggggg"},
		{"crimp", "toads", "bbbbb"},
		{"geese", "those", "bbbgg"},
		{"llama", "alley", "ygybb"},
		{"speed", "erase", "ybyyb"},
	}

	for _, tc := range tests {
		got, err := decodePattern(tc.pattern)
		if err != nil {
			t.Errorf("decodePattern(%q): %v", tc.pattern, err)
			continue
		}
		if want := solver.GetPattern(tc.guess, tc.answer); got != want {
			t.Errorf("decodePattern(%q) = %d, GetPattern(%q, %q) = %d",
				tc.pattern, got, tc.guess, tc.answer, want)
		}
	}
}

func TestDecodePatternRejectsBadInput(t *testing.T) {
	for _, s := range []string{"", "bbbb", "bbbbbb", "bbxbb", "ggggg ", "12345"} {
		if _, err := decodePattern(s); err == nil {
			t.Errorf("decodePattern(%q) succeeded, want an error", s)
		}
	}
}

func TestDecodePatternIgnoresCase(t *testing.T) {
	lower, err := decodePattern("bygbg")
	if err != nil {
		t.Fatal(err)
	}
	upper, err := decodePattern("BYGBG")
	if err != nil {
		t.Fatal(err)
	}
	if lower != upper {
		t.Errorf("case changed the pattern: %d vs %d", lower, upper)
	}
}

func TestDecodeMode(t *testing.T) {
	tests := []struct {
		in   string
		want solver.Mode
		ok   bool
	}{
		{"", solver.Easy, true},
		{"easy", solver.Easy, true},
		{"Hard", solver.Hard, true},
		{" hard ", solver.Hard, true},
		{"medium", 0, false},
	}

	for _, tc := range tests {
		got, err := decodeMode(tc.in)
		if tc.ok && err != nil {
			t.Errorf("decodeMode(%q): %v", tc.in, err)
		}
		if !tc.ok && err == nil {
			t.Errorf("decodeMode(%q) succeeded, want an error", tc.in)
		}
		if tc.ok && got != tc.want {
			t.Errorf("decodeMode(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

func TestDecodeHistoryRejectsOverlongHistory(t *testing.T) {
	history := make([]turnRequest, maxHistory+1)
	for i := range history {
		history[i] = turnRequest{Guess: "crane", Pattern: "bbbbb"}
	}
	if _, err := decodeHistory(history); err == nil {
		t.Fatal("want an error past the history limit")
	}
}

// The flag on the wire is the only way to reach the wider answer list, so its
// absence has to leave the engine on the official one. A client that has never
// heard of going off the books must not be able to end up there.
func TestDecodeOptionsCarriesTheUniverse(t *testing.T) {
	tests := []struct {
		offBook bool
		want    solver.Universe
	}{
		{false, solver.Official},
		{true, solver.OffBook},
	}

	for _, tc := range tests {
		opts, err := decodeOptions("easy", tc.offBook, nil, 5)
		if err != nil {
			t.Fatalf("decodeOptions(offBook=%v): %v", tc.offBook, err)
		}
		if opts.Universe != tc.want {
			t.Errorf("offBook=%v gave universe %d, want %d", tc.offBook, opts.Universe, tc.want)
		}
	}
}
