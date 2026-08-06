/**
 * A guess graded out of five stars.
 *
 * The rating is the share of the available information the guess actually
 * captured, not its rank in the pool. Rank reads badly here: with ~13,000
 * legal words the fifth-best opener is indistinguishable from the best in
 * play, yet ranking would show that as a whole star's difference. A ratio says
 * something the player can act on — "that was worth 84% of the best guess
 * going" — and keeps its meaning as the pool collapses underneath it.
 */
export const MAX_STARS = 5;

/**
 * bits is what the guess is worth, bestBits what the best guess in the same
 * position was worth. Never returns less than one star: a legal guess always
 * tells you something, and a scale that bottoms out at zero would make a poor
 * guess look like a broken one.
 */
export function rate(bits: number, bestBits: number): number {
	// Nothing left to learn, which happens once a single answer survives. Any
	// legal word is then as good as any other, so none of them is a bad play.
	if (bestBits <= 0) return MAX_STARS;

	const ratio = Math.max(0, bits) / bestBits;
	return Math.min(MAX_STARS, Math.max(1, MAX_STARS * ratio));
}

export type Tier = "great" | "good" | "plain";

/** The three bands the rating is coloured in. */
export function tier(stars: number): Tier {
	if (stars >= 4.5) return "great";
	if (stars >= 3.5) return "good";
	return "plain";
}

/** One decimal, so 5 reads as "5.0" and lines up with the ratings around it. */
export function format(stars: number): string {
	return stars.toFixed(1);
}
