import type { Color, Turn } from "./types";

export type Key = { label: string; value: string; wide?: boolean };

const letters = (row: string): Key[] =>
	row.split("").map((c) => ({ label: c, value: c }));

/** The real game's layout, including the two wide action keys. */
export const KEY_ROWS: Key[][] = [
	letters("qwertyuiop"),
	letters("asdfghjkl"),
	[
		{ label: "enter", value: "Enter", wide: true },
		...letters("zxcvbnm"),
		{ label: "⌫", value: "Backspace", wide: true },
	],
];

/** Better colours win, so a letter that ever went green stays green. */
const rank: Record<Color, number> = { absent: 0, present: 1, correct: 2 };

/**
 * Derives each key's colour from the feedback given so far. Like the candidate
 * set, this is recomputed from the whole history rather than accumulated, so
 * undoing a turn cleans up after itself with no extra bookkeeping.
 */
export function keyColors(history: Turn[]): Record<string, Color> {
	const colors: Record<string, Color> = {};

	for (const turn of history) {
		for (let i = 0; i < turn.guess.length; i++) {
			const letter = turn.guess[i];
			const color: Color =
				turn.pattern[i] === "g"
					? "correct"
					: turn.pattern[i] === "y"
						? "present"
						: "absent";

			const current = colors[letter];
			if (current === undefined || rank[color] > rank[current]) {
				colors[letter] = color;
			}
		}
	}

	return colors;
}
