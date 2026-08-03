// Package solver scores wordle guesses against the set of answers that are
// still possible.
//
// The engine never knows the answer. Feedback for a real guess is observed and
// relayed in by the caller; GetPattern only ever produces hypothetical
// feedback, used to bucket candidates while scoring and to test a past guess
// for consistency.
package solver

import "github.com/mmrzz/wordle-coach/internal/data"

// WordLen is the length of every word the engine handles.
const WordLen = data.WordLen

// Pattern is the feedback for one guess, packed as a base-3 code: digit i (the
// 3^i place) holds position i as Grey, Yellow or Green. Every pattern fits in
// 0..242, so a [NumPatterns]int array indexes them directly.
type Pattern uint8

// NumPatterns is the number of distinct patterns, 3^WordLen.
const NumPatterns = 243

// Feedback for a single position.
const (
	Grey   = 0 // letter is not in the answer, or its copies are spoken for
	Yellow = 1 // letter is in the answer, but not here
	Green  = 2 // letter is in the answer, here
)

// AllGreen is the winning pattern, every position Green.
const AllGreen Pattern = NumPatterns - 1

// Digit returns the feedback at position i.
func (p Pattern) Digit(i int) int {
	for ; i > 0; i-- {
		p /= 3
	}
	return int(p % 3)
}

// GetPattern returns the feedback real wordle would show for guess against
// answer. Both must be lowercase words of WordLen letters.
//
// This is the single source of truth for the whole engine: it defines what a
// score means and, by way of Filter, what "still possible" means. A bug here
// corrupts both at once.
func GetPattern(guess, answer string) Pattern {
	var digits [WordLen]uint8

	// Greens are resolved first, and the letters they consume are withheld
	// from the count, so a second copy of a letter can only go yellow if the
	// answer really has a second copy left over.
	var unmatched [26]int8
	for i := 0; i < WordLen; i++ {
		if guess[i] == answer[i] {
			digits[i] = Green
			continue
		}
		unmatched[answer[i]-'a']++
	}

	for i := 0; i < WordLen; i++ {
		if digits[i] == Green {
			continue
		}
		if c := guess[i] - 'a'; unmatched[c] > 0 {
			digits[i] = Yellow
			unmatched[c]--
		}
	}

	code := Pattern(0)
	for i := WordLen - 1; i >= 0; i-- {
		code = code*3 + Pattern(digits[i])
	}
	return code
}
