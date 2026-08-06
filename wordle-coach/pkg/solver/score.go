package solver

import "math"

// Beta is the Rényi order used to collapse a pattern histogram into a score.
// It is the one knob the temperature slider turns, and every scoring rule is a
// point on it:
//
//	β → 0   log2 of the number of distinct patterns. Bucket sizes are ignored,
//	        so it rewards splitting the field as many ways as possible however
//	        lopsided the split. Bold.
//	β = 1   Shannon entropy, the expected information gained. Balanced, and
//	        the default.
//	β = 2   equivalent to minimising Σnᵢ²/N, the expected number of answers
//	        left. Penalises big buckets. Safe.
//	β → ∞   approaches minimax: only the largest bucket counts, so it plays to
//	        make the worst case as small as possible. Safest.
//
// Higher scores are better at every β, so one comparison rule serves them all.
const (
	// DefaultBeta is Shannon entropy.
	DefaultBeta = 1.0
	// MinBeta is just above zero. Zero itself is rejected so that the zero
	// value of Options fails loudly instead of silently rescoring.
	MinBeta = 0.01
	// MaxBeta is far enough out to be effectively minimax.
	MaxBeta = 10.0
)

// betaEpsilon is how close to 1 counts as 1. The Rényi formula divides by
// (1-β), so near 1 it is numerically useless and the Shannon limit is used.
const betaEpsilon = 1e-6

// collapse reduces one histogram to its score and its expected remaining count.
//
// Both come off the same buckets in a single pass. ExpRemaining is reported at
// every β, not just β=2, because it is a plain description of the position
// rather than the objective being optimised.
func collapse(counts *[NumPatterns]int, total, beta float64) (bits, expRemaining float64) {
	if math.Abs(beta-DefaultBeta) < betaEpsilon {
		for _, n := range counts {
			if n == 0 {
				continue
			}
			p := float64(n) / total
			bits -= p * math.Log2(p)
			expRemaining += p * float64(n)
		}
		return bits, expRemaining
	}

	sum := 0.0
	for _, n := range counts {
		if n == 0 {
			continue
		}
		p := float64(n) / total
		sum += math.Pow(p, beta)
		expRemaining += p * float64(n)
	}

	// H_β = log2(Σ pᵢ^β) / (1-β). Positive at every β: below 1 the sum
	// exceeds 1 and both terms are positive, above 1 both are negative.
	return math.Log2(sum) / (1 - beta), expRemaining
}
