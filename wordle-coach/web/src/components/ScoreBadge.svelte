<script lang="ts">
	import type { Score } from "../lib/game.svelte";

	type Props = { score: Score | null; pending?: boolean };
	let { score, pending = false }: Props = $props();

	/*
	 * Rank is shown against the whole pool, so "31" out of nearly 13,000 is a
	 * fine guess even though the number looks large. The colour carries that
	 * meaning at a glance; the title attribute spells it out.
	 */
	const quality = $derived.by(() => {
		if (score?.percentile === undefined) return "plain";
		if (score.percentile >= 99.5) return "great";
		if (score.percentile >= 95) return "good";
		return "plain";
	});

	const title = $derived.by(() => {
		if (!score) return "";
		const parts = [`${score.bits.toFixed(2)} bits of information`];
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
	{#if score}
		<span class="bits">{score.bits.toFixed(2)}</span>
		{#if score.rank !== undefined}
			<span class="rank" data-quality={quality}>#{score.rank.toLocaleString()}</span>
		{:else}
			<span class="unit">bits</span>
		{/if}
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

	.bits {
		font-size: 0.95rem;
		font-weight: 700;
		color: var(--fg);
	}

	.unit,
	.rank,
	.empty,
	.dots {
		font-size: 0.68rem;
		color: var(--fg-muted);
	}

	.rank[data-quality="great"] {
		color: var(--color-correct);
		font-weight: 700;
	}

	.rank[data-quality="good"] {
		color: var(--darkened-yellow);
		font-weight: 700;
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
		.bits {
			font-size: 0.8rem;
		}
		.unit,
		.rank,
		.empty,
		.dots {
			font-size: 0.6rem;
		}
	}
</style>
