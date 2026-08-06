/**
 * Animation timings, matching the real game's.
 *
 * These are duplicated as custom properties in app.css because CSS needs them
 * for the keyframes and JS needs them to know when a reveal has finished. The
 * two lists have to be changed together.
 */
export const POP_MS = 100;
export const FLIP_MS = 500;
export const FLIP_STAGGER_MS = 300;
export const SHAKE_MS = 600;
export const BOUNCE_MS = 1000;
export const BOUNCE_STAGGER_MS = 100;
export const TOAST_MS = 1600;

/**
 * The suggestion landing: the same turn as a colour change, taken briskly.
 *
 * A colour reveal is the drama of the game and is paced for it. A suggestion
 * arriving is not — it is an answer to a question, and at the reveal's tempo
 * the wait for it would be longer than the wait that caused it.
 */
export const SETTLE_MS = 260;
export const SETTLE_STAGGER_MS = 60;

/**
 * How long the wait must last before it is worth admitting to.
 *
 * Most suggestions arrive well inside this, and a loading state that appears
 * for a tenth of a second reads as a glitch rather than as progress.
 */
export const SHIMMER_DELAY_MS = 180;
export const SHIMMER_STAGGER_MS = 60;

/** How long a whole row takes to turn over, including the stagger. */
export const ROW_REVEAL_MS = FLIP_MS + FLIP_STAGGER_MS * 4;

/** The same, for the quicker turn a landing suggestion takes. */
export const ROW_SETTLE_MS = SETTLE_MS + SETTLE_STAGGER_MS * 4;

const query =
	typeof matchMedia === "function"
		? matchMedia("(prefers-reduced-motion: reduce)")
		: null;

/**
 * Whether the viewer asked for less movement. Read at call time rather than
 * cached, so changing the system setting takes effect without a reload.
 */
export function prefersReducedMotion(): boolean {
	return query?.matches ?? false;
}

/**
 * Collapses a duration to nothing when reduced motion is requested. State
 * changes still happen, they just stop being animated, so nothing that depends
 * on an animation finishing can stall.
 */
export function duration(ms: number): number {
	return prefersReducedMotion() ? 0 : ms;
}
