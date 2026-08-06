<script lang="ts">
	import { untrack } from "svelte";
	import {
		FLIP_MS,
		POP_MS,
		SETTLE_MS,
		SETTLE_STAGGER_MS,
		SHIMMER_DELAY_MS,
		SHIMMER_STAGGER_MS,
		duration,
	} from "../lib/anim";
	import type { Color } from "../lib/types";

	type Props = {
		letter: string;
		/** null while the tile has not been turned over. */
		color: Color | null;
		/** True when the letter is our suggestion rather than the player's. */
		ghost?: boolean;
		/** Set during colouring, when clicking the tile changes its colour. */
		interactive?: boolean;
		/** Staggers the turn so a row reveals left to right. */
		flipDelay?: number;
		/** True while this tile is waiting for a word rather than holding one. */
		loading?: boolean;
		/** True when a letter arriving should turn in rather than pop in. */
		arriving?: boolean;
		position?: number;
		onclick?: () => void;
	};

	let {
		letter,
		color,
		ghost = false,
		interactive = false,
		flipDelay = 0,
		loading = false,
		arriving = false,
		position = 0,
		onclick,
	}: Props = $props();

	/*
	 * What is painted, as opposed to what is true. The two part company for
	 * half of the flip: the tile keeps its old colour until it is edge-on, then
	 * takes the new one on the way back up. Driving the swap from a timer
	 * rather than from an animation event keeps the tile in step with the
	 * stagger and means reduced motion collapses it to an instant change.
	 */
	// svelte-ignore state_referenced_locally
	// Seeded from the prop deliberately: a tile that mounts already coloured,
	// as when a saved game is restored, must not play a turn on arrival.
	let shown = $state<Color | null>(color);
	let flipping = $state(false);
	let popping = $state(false);
	let hasRendered = false;

	$effect(() => {
		const next = color;

		return untrack(() => {
			if (!hasRendered) {
				hasRendered = true;
				shown = next;
				return;
			}
			if (next === shown) return;

			const half = duration(FLIP_MS) / 2;
			flipping = true;

			const swap = setTimeout(() => (shown = next), flipDelay + half);
			const done = setTimeout(
				() => (flipping = false),
				flipDelay + duration(FLIP_MS),
			);

			return () => {
				clearTimeout(swap);
				clearTimeout(done);
			};
		});
	});

	/*
	 * The letter actually painted, which parts company with the prop only while
	 * a suggestion is landing. Then it arrives at the midpoint of the turn, when
	 * the tile is edge-on, exactly as a colour does.
	 */
	// svelte-ignore state_referenced_locally
	// Seeded from the prop for the same reason as the colour: a tile that mounts
	// with a letter already in it must not play its arrival.
	let shownLetter = $state(letter);
	let landing = $state(false);

	$effect(() => {
		const next = letter;
		const isGhost = ghost;
		const turns = arriving;

		return untrack(() => {
			if (next === shownLetter) return;
			const appeared = next !== "" && shownLetter === "";

			if (appeared && turns) {
				const delay = duration(position * SETTLE_STAGGER_MS);
				const half = duration(SETTLE_MS) / 2;
				landing = true;

				const swap = setTimeout(() => (shownLetter = next), delay + half);
				const done = setTimeout(() => (landing = false), delay + duration(SETTLE_MS));
				return () => {
					clearTimeout(swap);
					clearTimeout(done);
				};
			}

			shownLetter = next;

			// The letter's own little pop, skipped for our suggestion so the
			// prefilled row appears rather than types itself in.
			if (!appeared || isGhost) return;
			popping = true;
			const timer = setTimeout(() => (popping = false), duration(POP_MS));
			return () => clearTimeout(timer);
		});
	});

	/*
	 * One delay for whichever animation is running. They are mutually exclusive:
	 * a tile is either waiting for a word, taking one, or turning a colour over.
	 */
	const animationDelay = $derived.by(() => {
		if (loading) return `${SHIMMER_DELAY_MS + position * SHIMMER_STAGGER_MS}ms`;
		if (landing) return `${position * SETTLE_STAGGER_MS}ms`;
		if (flipping) return `${flipDelay}ms`;
		return undefined;
	});

	const label = $derived.by(() => {
		if (shownLetter === "") return "";
		const state = shown ?? (ghost ? "suggested" : "entered");
		const suffix = interactive ? ", click to change" : "";
		return `${shownLetter.toUpperCase()}, ${state}${suffix}`;
	});
</script>

<svelte:element
	this={interactive ? "button" : "div"}
	class="tile"
	class:filled={shownLetter !== "" && shown === null}
	class:ghost={ghost && shownLetter !== "" && shown === null}
	class:loading
	class:absent={shown === "absent"}
	class:present={shown === "present"}
	class:correct={shown === "correct"}
	class:flipping
	class:landing
	class:popping
	style:animation-delay={animationDelay}
	type={interactive ? "button" : undefined}
	role={interactive ? undefined : "img"}
	aria-label={label || undefined}
	aria-hidden={shownLetter === "" ? "true" : undefined}
	{onclick}
>
	<span aria-hidden="true">{shownLetter}</span>
</svelte:element>

<style>
	.tile {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 100%;
		height: 100%;
		font-size: clamp(1.5rem, 7vw, 2rem);
		font-weight: 700;
		line-height: 1;
		text-transform: uppercase;
		user-select: none;

		background: transparent;
		color: var(--fg);
		border: 2px solid var(--tile-border-empty);

		/* Perspective on the tile itself, so each one turns about its own axis. */
		transform: perspective(600px);
	}

	.tile.filled {
		border-color: var(--tile-border-filled);
	}

	/*
	 * Our suggestion, held lightly: present enough to accept with Enter,
	 * faint enough that it never reads as something the player typed.
	 */
	.tile.ghost {
		color: var(--fg-faint);
		border-color: var(--tile-border-empty);
		border-style: dashed;
	}

	.tile.absent,
	.tile.present,
	.tile.correct {
		color: var(--tile-text-evaluated);
		border-color: transparent;
	}

	.tile.absent {
		background: var(--color-absent);
	}
	.tile.present {
		background: var(--color-present);
	}
	.tile.correct {
		background: var(--color-correct);
	}

	.tile.popping {
		animation: PopIn var(--pop-ms) ease-in-out;
	}

	.tile.flipping {
		animation: Flip var(--flip-ms) ease-in;
	}

	/* The same turn as a colour change, taken at the quicker tempo. */
	.tile.landing {
		animation: Flip var(--settle-ms) ease-in;
	}

	/*
	 * The wait, and only once it has lasted long enough to be worth admitting
	 * to. The sweep begins off the tile, so the delay shows nothing at all and
	 * a suggestion that arrives quickly never flashes anything.
	 */
	.tile.loading {
		background-image: linear-gradient(
			100deg,
			transparent 25%,
			color-mix(in srgb, var(--fg) 14%, transparent) 50%,
			transparent 75%
		);
		background-size: 300% 100%;
		background-repeat: no-repeat;
		animation: Shimmer var(--shimmer-ms) ease-in-out infinite backwards;
	}

	/*
	 * An endless sweep collapsed to 1ms is a strobe, so the wait is told as a
	 * still tint instead of not being told at all.
	 */
	@media (prefers-reduced-motion: reduce) {
		.tile.loading {
			animation: none !important;
			background-image: none;
			background-color: color-mix(in srgb, var(--fg) 8%, transparent);
		}
	}

	/* Only the colouring pass makes tiles clickable. */
	button.tile {
		cursor: pointer;
		padding: 0;
		font-family: inherit;
	}

	button.tile:hover {
		filter: brightness(1.1);
	}

	button.tile:active {
		transform: perspective(600px) scale(0.96);
	}
</style>
