<script lang="ts">
	import type { Score } from "../lib/game.svelte";
	import { format, rate, tier } from "../lib/stars";

	type Props = { score: Score | null; pending?: boolean };
	let { score, pending = false }: Props = $props();

	/*
	 * Out of five, so a glance is enough. Bits and rank are the honest numbers
	 * underneath and stay in the tooltip: "31st of 12,972" is a fine guess
	 * despite looking like a large number, which is exactly why it is the wrong
	 * thing to put on the board.
	 */
	const stars = $derived(
		score?.bestBits === undefined ? null : rate(score.bits, score.bestBits),
	);

	const title = $derived.by(() => {
		if (!score) return "";

		const parts: string[] = [];
		if (stars !== null) parts.push(`${format(stars)} out of 5`);
		parts.push(`${score.bits.toFixed(2)} bits of information`);

		if (score.rank !== undefined && score.poolSize !== undefined) {
			parts.push(`rank ${score.rank.toLocaleString()} of ${score.poolSize.toLocaleString()}`);
		} else if (score.rank !== undefined) {
			parts.push(`rank ${score.rank}`);
		}
		if (score.percentile !== undefined) {
			parts.push(`better than ${score.percentile.toFixed(1)}% of legal guesses`);
		}
		return parts.join(" · ");
	});
</script>

<div class="badge" class:pending {title}>
	{#if score && stars !== null}
		<span class="stars" data-tier={tier(stars)}>
			{format(stars)}<span class="star" aria-hidden="true">★</span>
			<span class="sr-only">out of 5</span>
		</span>
		{#if score.rank !== undefined}
			<span class="rank">#{score.rank.toLocaleString()}</span>
		{/if}
	{:else if score}
		<!-- A save written before ratings existed has nothing to measure
		     against, so the raw number is all there is to show. -->
		<span class="stars">{score.bits.toFixed(2)}</span>
		<span class="rank">bits</span>
	{:else if pending}
		<span class="dots" aria-label="scoring">···</span>
	{:else}
		<span class="empty">—</span>
	{/if}
</div>

<style>
	.badge {
		display: flex;
		flex-direction: column;
		align-items: flex-start;
		justify-content: center;
		min-width: 3rem;
		font-variant-numeric: tabular-nums;
		line-height: 1.15;
	}

	.stars {
		font-size: 0.95rem;
		font-weight: 700;
		color: var(--fg);
	}

	.star {
		margin-left: 0.12em;
		font-size: 0.8em;
	}

	.stars[data-tier="great"] {
		color: var(--color-correct);
	}

	.stars[data-tier="good"] {
		color: var(--darkened-yellow);
	}

	.rank,
	.empty,
	.dots {
		font-size: 0.68rem;
		color: var(--fg-muted);
	}

	.empty,
	.dots {
		color: var(--fg-faint);
	}

	.pending .dots {
		animation: FadeUp 600ms ease-in-out infinite alternate;
	}

	@media (max-width: 620px) {
		.badge {
			min-width: 2.6rem;
		}
		.stars {
			font-size: 0.8rem;
		}
		.rank,
		.empty,
		.dots {
			font-size: 0.6rem;
		}
	}
</style>
