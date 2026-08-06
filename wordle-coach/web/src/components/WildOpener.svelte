<script lang="ts">
	import type { Game } from "../lib/game.svelte";
	import { randomWild } from "../lib/wild";

	type Props = { game: Game };
	let { game }: Props = $props();

	/*
	 * The card is only ever mounted on an empty board, so the word is drawn once
	 * per game and the reroll is the only thing that changes it. Picking again on
	 * every keystroke would swap the offer out from under the player's cursor.
	 */
	let pick = $state(randomWild());
	let chosen = $state(false);

	/** True while the offer is on the board: taken, and not since typed over. */
	const taken = $derived(chosen && (game.phase !== "typing" || game.activeWord === pick.word));

	const open = $derived(game.phase === "typing");

	function reroll() {
		pick = randomWild(pick.word);
		chosen = false;
	}

	function take() {
		game.apply(pick.word);
		chosen = true;
	}
</script>

{#if game.atStart}
	<section class="wild" aria-label="A weird opening word">
		<p class="ask">
			{taken ? "Wild it is." : "Want to go wild?"}
			<button
				class="reroll"
				onclick={reroll}
				disabled={!open}
				title="Another odd word"
				aria-label="Offer a different weird word"
			>
				↻
			</button>
		</p>

		<button
			class="offer"
			class:taken
			onclick={take}
			disabled={!open}
			title={`Put ${pick.word.toUpperCase()} in the row`}
		>
			Start with <strong>{pick.word}</strong> — it means {pick.meaning}.
		</button>
	</section>
{/if}

<style>
	.wild {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 2px;
		/* Tight underneath, so the offer reads as belonging to the board it sits
		   on rather than to the count above it. */
		padding: 10px 12px 6px;
		text-align: center;
		flex: none;
		animation: SlideUp 200ms ease-out;
	}

	.ask {
		display: flex;
		align-items: center;
		gap: 6px;
		margin: 0;
		font-size: 0.72rem;
		font-weight: 700;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		color: var(--fg-faint);
	}

	.reroll {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 18px;
		height: 18px;
		border-radius: 50%;
		font-size: 0.82rem;
		line-height: 1;
		color: var(--fg-muted);
	}

	.reroll:hover:not(:disabled) {
		color: var(--fg);
		background: color-mix(in srgb, var(--fg) 10%, transparent);
	}

	/*
	 * A line of prose that happens to be clickable, so it is styled as text
	 * rather than as a control: the word inside it carries the affordance.
	 */
	.offer {
		max-width: 42ch;
		padding: 4px 8px;
		border-radius: 6px;
		font-size: 0.78rem;
		line-height: 1.45;
		color: var(--fg-muted);
		transition: background 120ms ease, color 120ms ease;
	}

	.offer:hover:not(:disabled) {
		color: var(--fg);
		background: color-mix(in srgb, var(--color-correct) 12%, transparent);
	}

	.offer strong {
		font-weight: 700;
		text-transform: uppercase;
		letter-spacing: 0.1em;
		color: var(--fg);
		border-bottom: 2px solid var(--color-correct);
	}

	/* Accepted, or no longer takeable: either way the line stops advertising
	   itself and reads as a caption for the word now in the row. */
	.taken strong,
	.offer:disabled strong {
		border-bottom-color: transparent;
	}

	.offer:disabled,
	.reroll:disabled {
		cursor: default;
	}

	/* A phone in landscape has no vertical room to spare, and the board comes
	   first. */
	@media (max-height: 620px) {
		.wild {
			display: none;
		}
	}
</style>
