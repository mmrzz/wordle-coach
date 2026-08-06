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

// offBookOpenerTable is the same ranking for a game whose answer may be any
// legal word.
//
// It is a different ranking, not a rescaling of the one above: TARES leads
// here and SOARE only places seventh, because the wider answer set rewards a
// different spread of letters. Live it would cost ~168M pattern evaluations,
// every guess against every guess, which is the most expensive question the
// engine can be asked and the least informative, since nothing is known yet.
var offBookOpenerTable = []Suggestion{
	{Word: "tares", Bits: 6.194053, ExpRemaining: 302.4917, InAnswerSet: true},
	{Word: "lares", Bits: 6.149919, ExpRemaining: 288.7382, InAnswerSet: true},
	{Word: "rales", Bits: 6.114343, ExpRemaining: 292.1061, InAnswerSet: true},
	{Word: "rates", Bits: 6.096243, ExpRemaining: 311.3554, InAnswerSet: true},
	{Word: "teras", Bits: 6.076619, ExpRemaining: 337.0384, InAnswerSet: true},
	{Word: "nares", Bits: 6.066831, ExpRemaining: 305.5487, InAnswerSet: true},
	{Word: "soare", Bits: 6.061395, ExpRemaining: 303.8309, InAnswerSet: true},
	{Word: "tales", Bits: 6.054988, ExpRemaining: 325.5948, InAnswerSet: true},
}

// openers returns the n best opening guesses for u, or as many as are
// tabulated. Every word of an off-book game is a possible answer, which is why
// that table is flagged InAnswerSet throughout.
func openers(n int, u Universe) []Suggestion {
	table := openerTable
	if u == OffBook {
		table = offBookOpenerTable
	}

	out := make([]Suggestion, min(n, len(table)))
	copy(out, table)
	return out
}
