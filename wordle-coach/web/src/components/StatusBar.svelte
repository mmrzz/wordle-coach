<script lang="ts">
	import type { Game } from "../lib/game.svelte";

	type Props = { game: Game };
	let { game }: Props = $props();

	const message = $derived.by(() => {
		switch (game.phase) {
			case "loading":
				return "Loading the word list…";
			case "offline":
				return game.fatal;
			case "coloring":
				return "Click the tiles to match the colours you were shown, then press Enter";
			case "won":
				// Deduced means the field collapsed to one word without it
				// having been played, so "solved in N" would be misleading.
				if (game.outcome === "deduced") {
					return `Only ${game.answer?.toUpperCase()} is left — that's the answer`;
				}
				return `Solved in ${game.history.length} — press the refresh icon for a new game`;
			case "lost":
				return game.remaining.length === 1
					? `Out of guesses. It was ${game.remaining[0].toUpperCase()}.`
					: "Out of guesses.";
			default:
				return null;
		}
	});
</script>

<div class="bar" class:warn={game.phase === "offline"}>
	{#if message}
		<span class="hint">{message}</span>
	{:else}
		<span class="count">
			<strong>{game.possibleCount.toLocaleString()}</strong>
			{game.possibleCount === 1 ? "answer left" : "answers left"}
		</span>
		{#if game.offBook}
			<!-- The count means something different now, and a much larger
			     number with no explanation reads as a bug. -->
			<span class="offbook" title="The answer may be any legal word, not only an official one">
				off the books
			</span>
		{/if}
		{#if game.thinking}
			<span class="thinking" aria-label="Scoring">·</span>
		{/if}
	{/if}
</div>

<style>
	.bar {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 6px;
		min-height: 26px;
		padding: 4px 12px;
		font-size: 0.78rem;
		color: var(--fg-muted);
		text-align: center;
		flex: none;
	}

	.count strong {
		color: var(--fg);
		font-size: 0.9rem;
		font-variant-numeric: tabular-nums;
	}

	.offbook {
		padding: 1px 6px;
		border: 1px solid var(--border);
		border-radius: 10px;
		font-size: 0.68rem;
		letter-spacing: 0.04em;
		text-transform: uppercase;
		color: var(--fg-muted);
		white-space: nowrap;
	}

	.warn {
		color: #d64545;
		font-weight: 700;
	}

	.thinking {
		animation: FadeUp 500ms ease-in-out infinite alternate;
		font-size: 1.4rem;
		line-height: 0;
	}
</style>
