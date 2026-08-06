package solver

// LetterOdds is where one letter stands across the answers still in play.
//
// It is the position described a letter at a time rather than a word at a
// time, which is what a player needs to work the answer out for themselves:
// the engine can say "E is in four answers out of five, usually last" without
// naming a word and taking the puzzle away.
type LetterOdds struct {
	// Letter is the lowercase letter being described.
	Letter string
	// Presence is the share of candidates containing it at least once, 0 to 1.
	// 1 means every answer still in play has it; 0 means it is ruled out.
	Presence float64
	// Position is the slot it sits in most often, counting from 0, or -1 when
	// no candidate contains it anywhere.
	Position int
	// PositionOdds is the share of all candidates holding it at Position. It
	// is deliberately not conditioned on the letter appearing, so it reads on
	// the same scale as Presence and can never exceed it.
	PositionOdds float64
}

// letterOdds measures the whole alphabet against candidates and returns all 26
// letters in order. Letters that cannot appear are kept rather than dropped,
// because "ruled out" is itself worth showing.
func letterOdds(candidates []string) []LetterOdds {
	var present [26]int
	var atPosition [26][WordLen]int
	var seen [26]bool

	for _, w := range candidates {
		for i := 0; i < WordLen; i++ {
			c := w[i] - 'a'
			atPosition[c][i]++
			// Counted once per word, so a word with two Es leaves E present in
			// one answer rather than two.
			if !seen[c] {
				seen[c] = true
				present[c]++
			}
		}
		for i := 0; i < WordLen; i++ {
			seen[w[i]-'a'] = false
		}
	}

	total := float64(len(candidates))
	out := make([]LetterOdds, 26)
	for c := range out {
		// Ties go to the earlier slot, which only matters for letters spread
		// evenly enough that the answer does not hinge on it.
		best, bestCount := -1, 0
		for i, n := range atPosition[c] {
			if n > bestCount {
				best, bestCount = i, n
			}
		}

		out[c] = LetterOdds{Letter: string(rune('a' + c)), Position: best}
		if total > 0 {
			out[c].Presence = float64(present[c]) / total
			out[c].PositionOdds = float64(bestCount) / total
		}
	}
	return out
}
