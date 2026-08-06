<script lang="ts">
	import { untrack } from "svelte";
	import {
		BOUNCE_MS,
		BOUNCE_STAGGER_MS,
		FLIP_STAGGER_MS,
		SHAKE_MS,
		duration,
	} from "../lib/anim";
	import type { Row } from "../lib/game.svelte";
	import ScoreBadge from "./ScoreBadge.svelte";
	import Tile from "./Tile.svelte";

	type Props = {
		row: Row;
		index: number;
		/** True for the row the player is working on. */
		active?: boolean;
		/** True while the player is setting this row's colours. */
		coloring?: boolean;
		/** Staggers this row's tiles, set only for the reveal after Enter. */
		staggered?: boolean;
		/** True while this row is waiting for a suggestion to arrive. */
		loading?: boolean;
		bouncing?: boolean;
		/** Changes to replay the rejection shake. */
		shakeNonce?: number;
		pendingScore?: boolean;
		onTileClick?: (position: number) => void;
		onAsk?: () => void;
		onSolved?: () => void;
	};

	let {
		row,
		index,
		active = false,
		coloring = false,
		staggered = false,
		loading = false,
		bouncing = false,
		shakeNonce = 0,
		pendingScore = false,
		onTileClick,
		onAsk,
		onSolved,
	}: Props = $props();

	let shaking = $state(false);

	// Only the active row answers a rejection, and only after the first render
	// so that restoring a saved game does not shake every row on load.
	// svelte-ignore state_referenced_locally
	// Seeded from the prop deliberately, so a row that mounts partway through a
	// game does not replay a shake that happened before it existed.
	let seenNonce = shakeNonce;
	$effect(() => {
		const nonce = shakeNonce;

		return untrack(() => {
			if (nonce === seenNonce) return;
			seenNonce = nonce;
			if (!active) return;

			shaking = true;
			const timer = setTimeout(() => (shaking = false), duration(SHAKE_MS));
			return () => clearTimeout(timer);
		});
	});
</script>

<div class="wrap">
	<div
		class="row"
		class:shaking
		class:invalid={shaking}
		role="group"
		aria-label={`Guess ${index + 1}`}
	>
		{#each row.letters as letter, position (position)}
			<div
				class="cell"
				class:bouncing
				style:animation-delay={bouncing ? `${position * BOUNCE_STAGGER_MS}ms` : undefined}
			>
				<Tile
					{letter}
					color={row.colors[position]}
					ghost={row.ghost}
					interactive={coloring}
					flipDelay={staggered ? position * FLIP_STAGGER_MS : 0}
					{loading}
					arriving={staggered}
					{position}
					onclick={() => onTileClick?.(position)}
				/>
			</div>
		{/each}
	</div>

	<div class="aside">
		<!-- Rows not yet reached carry no score, and an empty placeholder beside
		     each one would just be clutter. -->
		{#if row.score || pendingScore || active}
			<ScoreBadge score={row.score} pending={pendingScore} />
		{/if}
		<!--
			The two buttons swap by phase, which keeps the column one button
			wide. Asking for alternatives is pointless once the word is locked,
			and there is nothing to declare solved before it is.
		-->
		{#if coloring}
			<button
				class="tick"
				onclick={onSolved}
				aria-label="This guess was the answer, end the game"
				title="Got it — this was the answer"
			>
				<svg viewBox="0 0 24 24" width="16" height="16" aria-hidden="true">
					<path
						d="M5 12.5l4.5 4.5L19 7.5"
						fill="none"
						stroke="currentColor"
						stroke-width="2.6"
						stroke-linecap="round"
						stroke-linejoin="round"
					/>
				</svg>
			</button>
		{:else if active}
			<button
				class="ask"
				onclick={onAsk}
				aria-label="Show the five best guesses for this position"
				title="Best guesses for this position"
			>
				?
			</button>
		{/if}
	</div>
</div>

<style>
	.wrap {
		position: relative;
	}

	.row {
		display: grid;
		grid-template-columns: repeat(5, 1fr);
		gap: 5px;
	}

	.cell {
		aspect-ratio: 1;
	}

	.cell.bouncing {
		animation: Bounce var(--bounce-ms) ease-in-out;
	}

	.row.shaking {
		animation: Shake var(--shake-ms);
	}

	/*
	 * The requested red glow on rejection. It rides along with the shake so a
	 * bad word reads as wrong even with the sound off and the toast missed.
	 */
	.row.invalid :global(.tile) {
		border-color: #d64545;
		box-shadow: 0 0 8px rgba(214, 69, 69, 0.55);
	}

	/*
	 * Hung off the right edge without taking part in the layout, so the board
	 * itself stays exactly where the real game puts it.
	 */
	.aside {
		position: absolute;
		top: 0;
		bottom: 0;
		left: 100%;
		display: flex;
		align-items: center;
		gap: 6px;
		padding-left: 10px;
	}

	.ask,
	.tick {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 28px;
		height: 28px;
		border-radius: 50%;
		border: 2px solid var(--border-strong);
		color: var(--fg-muted);
		font-size: 0.95rem;
		font-weight: 700;
		flex: none;
		transition: background 120ms ease, color 120ms ease, transform 120ms ease;
	}

	.ask:hover,
	.tick:hover {
		background: var(--color-correct);
		border-color: var(--color-correct);
		color: #fff;
		transform: scale(1.08);
	}

	/* Green from the start: it is the winning action, not a neutral one. */
	.tick {
		border-color: var(--color-correct);
		color: var(--color-correct);
	}

	@media (max-width: 620px) {
		.aside {
			padding-left: 6px;
			gap: 4px;
		}
		.ask,
		.tick {
			width: 24px;
			height: 24px;
			font-size: 0.8rem;
		}
	}
</style>
