package solver

// openerTable is the top of the turn-one ranking, precomputed.
//
// With no feedback yet every game opens from the same position, so this ranking
// is a constant; computing it live would mean ~30M pattern evaluations to
// rediscover it on every new game. TestOpenerTableMatchesData keeps it honest
// and prints a replacement table if the word lists ever change.
var openerTable = []Suggestion{
	{Word: "soare", Bits: 5.885960, ExpRemaining: 62.3011, InAnswerSet: false},
	{Word: "roate", Bits: 5.882779, ExpRemaining: 60.4246, InAnswerSet: false},
	{Word: "raise", Bits: 5.877910, ExpRemaining: 61.0009, InAnswerSet: true},
	{Word: "raile", Bits: 5.865710, ExpRemaining: 61.3309, InAnswerSet: false},
	{Word: "reast", Bits: 5.865457, ExpRemaining: 71.7654, InAnswerSet: false},
	{Word: "slate", Bits: 5.855775, ExpRemaining: 71.5728, InAnswerSet: true},
	{Word: "crate", Bits: 5.834874, ExpRemaining: 72.8998, InAnswerSet: true},
	{Word: "salet", Bits: 5.834582, ExpRemaining: 71.2721, InAnswerSet: false},
}

// openers returns the n best opening guesses, or as many as are tabulated.
func openers(n int) []Suggestion {
	out := make([]Suggestion, min(n, len(openerTable)))
	copy(out, openerTable)
	return out
}
