import { WORD_LENGTH, type Color, type Turn } from "./types";

/**
 * Pre-colours a freshly locked row from what earlier turns already proved.
 *
 * A green fixes its letter to that square forever, so guessing the same letter
 * there again must come back green. A yellow proves the answer holds a copy of
 * that letter, so guessing it again is at least yellow. Starting every row grey
 * would make the player enter both facts a second time.
 *
 * It never colours a tile more strongly than the truth. Where it cannot prove
 * something it leaves the tile grey, so a wrong tile is always one the player
 * needs to raise, never one they need to take back.
 */
export function seedColors(history: Turn[], guess: string): Color[] {
	/** Squares proved to hold a particular letter. */
	const green: (string | null)[] = Array(WORD_LENGTH).fill(null);

	/**
	 * Copies of each letter the answer is known to hold, at minimum.
	 *
	 * The most any single turn showed coloured is the safe floor: a turn with
	 * two coloured Es proves two Es exist. Other turns cannot lower it, so the
	 * maximum is the strongest thing the history supports.
	 */
	const known = new Map<string, number>();

	for (const turn of history) {
		const coloured = new Map<string, number>();

		for (let i = 0; i < WORD_LENGTH; i++) {
			const letter = turn.guess[i];
			const code = turn.pattern[i];

			if (code === "g") green[i] = letter;
			if (code === "g" || code === "y") {
				coloured.set(letter, (coloured.get(letter) ?? 0) + 1);
			}
		}

		for (const [letter, count] of coloured) {
			known.set(letter, Math.max(known.get(letter) ?? 0, count));
		}
	}

	// A letter never seen coloured is absent, which the default already says.
	const colors: Color[] = Array(WORD_LENGTH).fill("absent");

	// Proved squares. Sound whatever else is going on, so they go down first.
	for (let i = 0; i < WORD_LENGTH; i++) {
		if (green[i] !== null && guess[i] === green[i]) colors[i] = "correct";
	}

	const occurrences = new Map<string, number>();
	for (const letter of guess) {
		occurrences.set(letter, (occurrences.get(letter) ?? 0) + 1);
	}

	/*
	 * A letter guessed no more times than the answer is known to contain it
	 * cannot come back grey anywhere: the game pairs each guessed copy with a
	 * distinct copy in the answer, and there are enough to go round. So every
	 * one of them is at least yellow, whichever squares they sit on.
	 *
	 * Guess the letter more often than that and the extra copies would come
	 * back grey, with nothing in the history saying which ones. Those are left
	 * alone rather than guessed at, since a tile coloured too strongly is worse
	 * than one left grey: it can be believed and confirmed without a second
	 * look, and a wrong pattern quietly corrupts every suggestion after it.
	 */
	for (let i = 0; i < WORD_LENGTH; i++) {
		if (colors[i] === "correct") continue;

		const letter = guess[i];
		if ((occurrences.get(letter) ?? 0) <= (known.get(letter) ?? 0)) {
			colors[i] = "present";
		}
	}

	return colors;
}
