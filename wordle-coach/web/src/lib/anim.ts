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

/** How long a whole row takes to turn over, including the stagger. */
export const ROW_REVEAL_MS = FLIP_MS + FLIP_STAGGER_MS * 4;

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
