/** The three tile states, named as the API spells them. */
export type Color = "absent" | "present" | "correct";

/** Click order for a tile: grey, then yellow, then green, then back to grey. */
export const COLOR_CYCLE: Color[] = ["absent", "present", "correct"];

/** Wire spelling of each colour, matching the Go pattern alphabet. */
export const COLOR_CODE: Record<Color, string> = {
	absent: "b",
	present: "y",
	correct: "g",
};

export type Mode = "easy" | "hard";

/** One played guess and the colours the real game showed for it. */
export type Turn = {
	guess: string;
	/** Five characters of g, y or b. */
	pattern: string;
};

export type Suggestion = {
	word: string;
	bits: number;
	expRemaining: number;
	inAnswerSet: boolean;
};

/**
 * Where one letter stands across the answers still in play. The engine sends
 * all 26 in alphabetical order, ruled-out ones included, and the grouping and
 * ordering are left to the UI.
 */
export type LetterOdds = {
	letter: string;
	/** Share of the remaining answers containing it, 0 to 1. */
	presence: number;
	/** The slot it sits in most often, counting from 0, or -1 if ruled out. */
	position: number;
	/** Share of all remaining answers holding it at that slot. */
	positionOdds: number;
};

export const VOWELS = new Set(["a", "e", "i", "o", "u"]);

export type SuggestResponse = {
	possibleCount: number;
	suggestions: Suggestion[];
	/** Only sent when few enough answers remain to list them. */
	remaining?: string[];
	letters: LetterOdds[];
};

export type RateResponse = {
	played: string;
	playedBits: number;
	playedExpRemaining: number;
	playedRank: number;
	percentile: number;
	best: Suggestion;
	gapBits: number;
	poolSize: number;
	possibleCount: number;
};

export type WordsResponse = {
	allowed: string[];
	answerCount: number;
};

export const WORD_LENGTH = 5;
export const MAX_TURNS = 6;

/** The pattern that means the puzzle is solved. */
export const WINNING_PATTERN = "ggggg";

/** Turns a row of colours into the wire pattern. */
export function toPattern(colors: Color[]): string {
	return colors.map((c) => COLOR_CODE[c]).join("");
}

/** Advances a tile one step around the colour cycle. */
export function nextColor(current: Color): Color {
	const i = COLOR_CYCLE.indexOf(current);
	return COLOR_CYCLE[(i + 1) % COLOR_CYCLE.length];
}
