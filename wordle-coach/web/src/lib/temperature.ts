/**
 * The temperature slider, which is the Renyi order beta the engine scores with.
 *
 * The slider track is log10(beta), so the range runs symmetrically from 0.1 to
 * 10 and the balanced default sits exactly in the middle rather than squashed
 * against the left edge, which is where a linear track would put it.
 */
export const DEFAULT_BETA = 1;
export const MIN_LOG = -1; // beta 0.1
export const MAX_LOG = 1; // beta 10
export const LOG_STEP = 0.05;

/** Slider position for a beta. */
export function toSlider(beta: number): number {
	return Math.log10(beta);
}

/** Beta for a slider position, rounded so the display never shows float noise. */
export function fromSlider(position: number): number {
	return Number(Math.pow(10, position).toPrecision(3));
}

export type Band = {
	name: string;
	blurb: string;
};

/**
 * What the current beta means, in the terms the slider is sold in. The
 * boundaries are presentational; the engine treats beta as continuous.
 */
export function band(beta: number): Band {
	if (beta < 0.5) {
		return {
			name: "Bold",
			blurb:
				"Chases the most possible outcomes, however uneven. Can split the field brilliantly or leave a big pile behind.",
		};
	}
	if (beta < 0.9) {
		return {
			name: "Curious",
			blurb: "Leans towards splitting the field wide over playing it safe.",
		};
	}
	if (beta <= 1.15) {
		return {
			name: "Balanced",
			blurb:
				"Plain information gain: the guess that tells you the most on average.",
		};
	}
	if (beta < 3) {
		return {
			name: "Safe",
			blurb:
				"Minimises how many answers are left afterwards. Punishes guesses with one big lopsided outcome.",
		};
	}
	return {
		name: "Cautious",
		blurb:
			"Plays against the worst case, making the largest group of survivors as small as it can.",
	};
}

/** Labelled stops drawn under the track. */
export const TICKS: { position: number; label: string }[] = [
	{ position: MIN_LOG, label: "Bold" },
	{ position: 0, label: "Balanced" },
	{ position: MAX_LOG, label: "Cautious" },
];
