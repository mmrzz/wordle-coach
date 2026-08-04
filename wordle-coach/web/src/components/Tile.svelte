<script lang="ts">
	import { untrack } from "svelte";
	import { FLIP_MS, POP_MS, duration } from "../lib/anim";
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
		position?: number;
		onclick?: () => void;
	};

	let {
		letter,
		color,
		ghost = false,
		interactive = false,
		flipDelay = 0,
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

	// The letter's own little pop as it lands, skipped for our suggestion so
	// the prefilled row appears rather than types itself in.
	let previousLetter = "";
	$effect(() => {
		const next = letter;
		const isGhost = ghost;

		return untrack(() => {
			const appeared = next !== "" && previousLetter === "";
			previousLetter = next;
			if (!appeared || isGhost) return;

			popping = true;
			const timer = setTimeout(() => (popping = false), duration(POP_MS));
			return () => clearTimeout(timer);
		});
	});

	const label = $derived.by(() => {
		if (letter === "") return "";
		const state = shown ?? (ghost ? "suggested" : "entered");
		const suffix = interactive ? ", click to change" : "";
		return `${letter.toUpperCase()}, ${state}${suffix}`;
	});
</script>

<svelte:element
	this={interactive ? "button" : "div"}
	class="tile"
	class:filled={letter !== "" && shown === null}
	class:ghost={ghost && shown === null}
	class:absent={shown === "absent"}
	class:present={shown === "present"}
	class:correct={shown === "correct"}
	class:flipping
	class:popping
	style:animation-delay={flipping ? `${flipDelay}ms` : undefined}
	type={interactive ? "button" : undefined}
	role={interactive ? undefined : "img"}
	aria-label={label || undefined}
	aria-hidden={letter === "" ? "true" : undefined}
	{onclick}
>
	<span aria-hidden="true">{letter}</span>
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
