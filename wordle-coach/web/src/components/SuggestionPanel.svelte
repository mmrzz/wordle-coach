<script lang="ts">
	import { definitionUrl } from "../lib/dictionary";
	import type { Game } from "../lib/game.svelte";
	import { format, rate, tier } from "../lib/stars";
	import { band } from "../lib/temperature";
	import LetterOdds from "./LetterOdds.svelte";
	import Modal from "./Modal.svelte";

	type Props = { game: Game };
	let { game }: Props = $props();

	const close = () => (game.panelOpen = false);

	// The list is the top of the pool, so its head is the best guess going and
	// every rating in the panel is measured against the same word.
	const bestBits = $derived(game.suggestions[0]?.bits ?? 0);
</script>

<Modal open={game.panelOpen} title="Best guesses" onclose={close}>
	<p class="lede">
		{#if game.thinking}
			Scoring…
		{:else}
			<strong>{game.possibleCount.toLocaleString()}</strong>
			{game.possibleCount === 1 ? "answer is" : "answers are"} still possible.
		{/if}
	</p>

	{#if game.remaining.length > 0}
		<p class="remaining">
			{#if game.remaining.length === 1}
				It has to be <strong>{game.remaining[0].toUpperCase()}</strong>.
			{:else}
				Only these are left:
				{#each game.remaining as word, i (word)}<span
						>{word.toUpperCase()}{i < game.remaining.length - 1 ? ", " : ""}</span
					>{/each}
			{/if}
		</p>
	{/if}

	<ol class="list">
		{#each game.suggestions as suggestion, i (suggestion.word)}
			<li>
				<button
					class="pick"
					onclick={() => game.apply(suggestion.word)}
					disabled={game.phase !== "typing"}
					title={`Put ${suggestion.word.toUpperCase()} in the row`}
				>
					<span class="rank">{i + 1}</span>
					<span class="word">{suggestion.word}</span>
					<span
						class="stats"
						title={`${suggestion.bits.toFixed(2)} bits of information`}
					>
						<span class="score" data-tier={tier(rate(suggestion.bits, bestBits))}>
							{format(rate(suggestion.bits, bestBits))}<span
								class="star"
								aria-hidden="true">★</span
							>
						</span>
						<span class="left">leaves ~{suggestion.expRemaining.toFixed(1)}</span>
					</span>
					{#if suggestion.inAnswerSet}
						<span class="could-win" title="This word could be the answer">✦</span>
					{/if}
				</button>
				<a
					class="define"
					href={definitionUrl(suggestion.word)}
					target="_blank"
					rel="noopener noreferrer"
					title={`Look up ${suggestion.word.toUpperCase()}`}
					aria-label={`Definition of ${suggestion.word.toUpperCase()}, opens in a new tab`}
				>
					↗
				</a>
			</li>
		{/each}
	</ol>

	<!-- On a wide screen this lives in the right-hand rail, where it can stay
	     open while you play. There is no room for a rail here. -->
	<div class="railed">
		<LetterOdds {game} />
	</div>

	<p class="footnote">
		<strong>★</strong> rates a guess out of five against the best one going, so
		five stars means nothing else would have told you more.
		<strong>leaves</strong> is how many answers you should expect to be left afterwards.
	</p>
	<p class="footnote">
		Click a word to play it, or <span class="arrow">↗</span> for its definition.
		Scored on <strong>{band(game.beta).name}</strong> — {band(game.beta).blurb.toLowerCase()}
	</p>
</Modal>

<style>
	.lede {
		margin: 0 0 10px;
		font-size: 0.9rem;
		color: var(--fg-muted);
	}

	.lede strong {
		color: var(--fg);
		font-size: 1.05rem;
	}

	.remaining {
		margin: 0 0 12px;
		padding: 8px 10px;
		border-radius: 6px;
		background: color-mix(in srgb, var(--color-correct) 14%, transparent);
		font-size: 0.82rem;
		letter-spacing: 0.03em;
	}

	.list {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: 6px;
	}

	li {
		display: flex;
		align-items: stretch;
		gap: 4px;
	}

	.pick {
		display: flex;
		align-items: center;
		gap: 10px;
		flex: 1;
		min-width: 0;
		padding: 8px 10px;
		border: 1px solid var(--border);
		border-radius: 6px;
		text-align: left;
		transition: border-color 120ms ease, background 120ms ease;
	}

	.pick:hover:not(:disabled) {
		border-color: var(--color-correct);
		background: color-mix(in srgb, var(--color-correct) 10%, transparent);
	}

	.pick:disabled {
		opacity: 0.55;
		cursor: default;
	}

	.rank {
		width: 1.1rem;
		font-size: 0.72rem;
		color: var(--fg-faint);
		font-variant-numeric: tabular-nums;
	}

	.word {
		font-size: 1.05rem;
		font-weight: 700;
		text-transform: uppercase;
		letter-spacing: 0.08em;
	}

	.stats {
		display: flex;
		flex-direction: column;
		margin-left: auto;
		text-align: right;
		font-variant-numeric: tabular-nums;
	}

	.score {
		font-size: 0.78rem;
		font-weight: 700;
	}

	.score[data-tier="great"] {
		color: var(--color-correct);
	}

	.score[data-tier="good"] {
		color: var(--darkened-yellow);
	}

	.star {
		margin-left: 0.12em;
		font-size: 0.85em;
	}

	/* Only when the rail it normally lives in has been hidden, so the two are
	   never on screen at once saying the same thing. */
	.railed {
		display: none;
		margin-top: 14px;
		padding-top: 12px;
		border-top: 1px solid var(--divider);
	}

	@media (max-width: 1060px) {
		.railed {
			display: block;
		}
	}

	.left {
		font-size: 0.66rem;
		color: var(--fg-muted);
	}

	.could-win {
		color: var(--color-correct);
		font-size: 0.9rem;
	}

	.define {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 34px;
		border: 1px solid var(--border);
		border-radius: 6px;
		color: var(--fg-muted);
		text-decoration: none;
		font-size: 0.9rem;
		flex: none;
	}

	.define:hover {
		color: var(--fg);
		border-color: var(--border-strong);
	}

	.footnote {
		margin: 12px 0 0;
		font-size: 0.72rem;
		line-height: 1.5;
		color: var(--fg-muted);
	}

	.footnote strong {
		color: var(--fg);
	}

	.arrow {
		color: var(--fg);
	}
</style>
