<script lang="ts">
	import { definitionUrl } from "../lib/dictionary";
	import type { Game } from "../lib/game.svelte";

	type Props = { game: Game };
	let { game }: Props = $props();

	const words = $derived(game.shortlist);
</script>

{#if words.length > 0}
	<section class="shortlist" aria-label="Answers still possible">
		<h2>{words.length === 1 ? "It has to be" : "Down to these"}</h2>

		<ul>
			{#each words as word (word)}
				<li>
					<button
						class="pick"
						onclick={() => game.apply(word)}
						disabled={game.phase !== "typing"}
						title={`Put ${word.toUpperCase()} in the row`}
					>
						{word}
					</button>
					<a
						class="define"
						href={definitionUrl(word)}
						target="_blank"
						rel="noopener noreferrer"
						aria-label={`Definition of ${word.toUpperCase()}, opens in a new tab`}
						title={`Look up ${word.toUpperCase()}`}
					>
						↗
					</a>
				</li>
			{/each}
		</ul>

		<p class="note">
			{#if words.length === 1}
				Nothing else fits the colours you have given.
			{:else}
				Every other answer is ruled out. Click one to play it.
			{/if}
		</p>
	</section>
{/if}

<style>
	.shortlist {
		padding: 12px;
		border: 1px solid var(--border);
		border-radius: 10px;
		background: var(--bg-raised);
		box-shadow: 0 2px 12px var(--shadow);
		animation: SlideUp 160ms ease-out;
	}

	h2 {
		margin: 0 0 8px;
		font-size: 0.72rem;
		font-weight: 700;
		letter-spacing: 0.1em;
		text-transform: uppercase;
		color: var(--fg-muted);
	}

	ul {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: 4px;
	}

	li {
		display: flex;
		gap: 4px;
	}

	.pick {
		flex: 1;
		min-width: 0;
		padding: 6px 8px;
		border: 1px solid var(--border);
		border-radius: 6px;
		text-align: left;
		font-size: 0.95rem;
		font-weight: 700;
		text-transform: uppercase;
		letter-spacing: 0.1em;
		transition: border-color 120ms ease, background 120ms ease;
	}

	.pick:hover:not(:disabled) {
		border-color: var(--color-correct);
		background: color-mix(in srgb, var(--color-correct) 12%, transparent);
	}

	.pick:disabled {
		opacity: 0.55;
		cursor: default;
	}

	.define {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 28px;
		border: 1px solid var(--border);
		border-radius: 6px;
		color: var(--fg-muted);
		text-decoration: none;
		font-size: 0.85rem;
		flex: none;
	}

	.define:hover {
		color: var(--fg);
		border-color: var(--border-strong);
	}

	.note {
		margin: 8px 0 0;
		font-size: 0.66rem;
		line-height: 1.45;
		color: var(--fg-faint);
	}
</style>
