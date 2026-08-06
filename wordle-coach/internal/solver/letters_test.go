package solver

import (
	"math"
	"testing"
)

// index returns the odds for one letter out of the alphabetical 26.
func index(odds []LetterOdds, letter string) LetterOdds {
	return odds[letter[0]-'a']
}

func TestLetterOddsCountsWordsNotLetters(t *testing.T) {
	// EERIE holds three Es. Presence asks whether the answer contains the
	// letter, so it must not count them.
	got := index(letterOdds([]string{"eerie", "slate"}), "e")

	if got.Presence != 1 {
		t.Errorf("Presence = %g, want 1: both words contain an E", got.Presence)
	}
	// E ends both words, and no other slot holds it twice.
	if got.Position != 4 {
		t.Errorf("Position = %d, want 4", got.Position)
	}
	if got.PositionOdds != 1 {
		t.Errorf("PositionOdds = %g, want 1", got.PositionOdds)
	}
}

func TestLetterOddsRulesOutMissingLetters(t *testing.T) {
	odds := letterOdds([]string{"slate", "crane"})

	got := index(odds, "z")
	if got.Presence != 0 || got.PositionOdds != 0 {
		t.Errorf("Z = %+v, want zero odds", got)
	}
	// -1 rather than 0, so "never seen" cannot be read as "usually first".
	if got.Position != -1 {
		t.Errorf("Position = %d, want -1 for a letter no candidate has", got.Position)
	}
}

func TestLetterOddsCoversTheAlphabetInOrder(t *testing.T) {
	odds := letterOdds([]string{"slate"})

	if len(odds) != 26 {
		t.Fatalf("got %d letters, want 26", len(odds))
	}
	for i, l := range odds {
		if want := string(rune('a' + i)); l.Letter != want {
			t.Fatalf("odds[%d].Letter = %q, want %q", i, l.Letter, want)
		}
	}
}

// PositionOdds is a share of every candidate, not of the ones holding the
// letter, so it can never exceed Presence. The whole alphabet is checked
// against the real answer list rather than a handful of words.
func TestPositionOddsNeverExceedPresence(t *testing.T) {
	e := testEngine(t)

	for _, l := range letterOdds(e.set.Answers) {
		if l.PositionOdds > l.Presence+1e-9 {
			t.Errorf("%s: PositionOdds %g exceeds Presence %g", l.Letter, l.PositionOdds, l.Presence)
		}
		if l.Presence < 0 || l.Presence > 1 {
			t.Errorf("%s: Presence %g is outside 0..1", l.Letter, l.Presence)
		}
	}
}

// A green letter is in every candidate and always in the same slot, which is
// the strongest statement the panel can make.
func TestLetterOddsAfterAGreen(t *testing.T) {
	e := testEngine(t)

	history := []Turn{{Guess: "slate", Pattern: GetPattern("slate", "sword")}}
	candidates := e.Filter(history)
	if len(candidates) == 0 {
		t.Fatal("no candidates survive, so there is nothing to measure")
	}
	got := index(letterOdds(candidates), "s")

	if math.Abs(got.Presence-1) > 1e-9 {
		t.Errorf("Presence = %g, want 1: every candidate starts with S", got.Presence)
	}
	if got.Position != 0 || math.Abs(got.PositionOdds-1) > 1e-9 {
		t.Errorf("position = %d at %g, want 0 at 1", got.Position, got.PositionOdds)
	}
}

// Both the opener fast path and the scored path have to carry the odds, and
// the numbers have to follow the feedback rather than the whole answer list.
func TestSuggestReportsLetterOdds(t *testing.T) {
	e := testEngine(t)

	t.Run("opening", func(t *testing.T) {
		got, err := e.Suggest(nil, easy(5))
		if err != nil {
			t.Fatal(err)
		}
		if len(got.Letters) != 26 {
			t.Fatalf("got %d letters, want 26", len(got.Letters))
		}
		if p := index(got.Letters, "e").Presence; p < 0.3 {
			t.Errorf("E presence = %g, want the commonest letter in a good share of answers", p)
		}
	})

	t.Run("mid game", func(t *testing.T) {
		// SLATE against PILOT shows a yellow L and a yellow T, so every
		// candidate has an L and none has an E.
		got, err := e.Suggest([]Turn{{Guess: "slate", Pattern: GetPattern("slate", "pilot")}}, easy(5))
		if err != nil {
			t.Fatal(err)
		}
		if p := index(got.Letters, "l").Presence; p != 1 {
			t.Errorf("L presence = %g, want 1: it came back yellow", p)
		}
		if p := index(got.Letters, "e").Presence; p != 0 {
			t.Errorf("E presence = %g, want 0: it came back grey", p)
		}
		// Yellow at slot 1 means L is anywhere but there.
		if got.Letters['l'-'a'].Position == 1 {
			t.Error("L cannot most likely sit where it was already shown yellow")
		}
	})
}
