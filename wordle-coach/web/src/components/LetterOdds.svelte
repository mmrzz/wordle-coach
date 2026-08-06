<script lang="ts">
	import type { Game } from "../lib/game.svelte";
	import { VOWELS, type LetterOdds } from "../lib/types";

	type Props = { game: Game };
	let { game }: Props = $props();

	/*
	 * Likeliest first rather than alphabetical. The list is here to answer
	 * "what should I try next", and A-to-Z buries the answer to that among
	 * twenty letters that have already been ruled out.
	 */
	function rank(letters: LetterOdds[]): LetterOdds[] {
		return [...letters].sort(
			(a, b) => b.presence - a.presence || a.letter.localeCompare(b.letter),
		);
	}

	const vowels = $derived(rank(game.letters.filter((l) => VOWELS.has(l.letter))));
	const consonants = $derived(rank(game.letters.filter((l) => !VOWELS.has(l.letter))));

	const ORDINALS = ["1st", "2nd", "3rd", "4th", "5th"];

	function percent(share: number): string {
		if (share === 0) return "0%";
		// Anything that survives at all is worth distinguishing from nothing,
		// and anything short of certain from certain.
		if (share < 0.01) return "<1%";
		if (share > 0.99 && share < 1) return ">99%";
		return `${Math.round(share * 100)}%`;
	}

	/**
	 * How much of the letter's showing is at its best slot. A letter in every
	 * answer but scattered across all five places has nothing useful to say
	 * about position, and should not look as if it does.
	 */
	function certainty(l: LetterOdds): number {
		return l.presence > 0 ? l.positionOdds / l.presence : 0;
	}

	function describe(l: LetterOdds): string {
		const name = l.letter.toUpperCase();
		if (l.presence === 0) return `${name} is ruled out`;

		const where =
			l.position >= 0
				? `, ${Math.round(certainty(l) * 100)}% of those in the ${ORDINALS[l.position]} spot`
				: "";
		return `${name} is in ${percent(l.presence)} of the answers still possible${where}`;
	}
</script>

{#snippet group(title: string, letters: LetterOdds[])}
	<h3>{title}</h3>
	<ul>
		{#each letters as l (l.letter)}
			<li class:out={l.presence === 0} title={describe(l)}>
				<span class="letter">{l.letter}</span>
				<span class="track" aria-hidden="true">
					<span class="fill" style:width={`${l.presence * 100}%`}></span>
				</span>
				<span class="pct">{percent(l.presence)}</span>
				<span class="spot" class:vague={certainty(l) < 0.5}>
					{l.position >= 0 ? ORDINALS[l.position] : "—"}
				</span>
			</li>
		{/each}
	</ul>
{/snippet}

<section class="odds" aria-label="Letter odds">
	<h2>Letters</h2>
	<p class="lede">
		Chance each one is in the answer, and where it most likely sits. Enough to
		work it out yourself.
	</p>

	{#if game.letters.length === 0}
		<p class="lede">Waiting for the solver…</p>
	{:else}
		{@render group("Vowels", vowels)}
		{@render group("Consonants", consonants)}
	{/if}
</section>

<style>
	/*
	 * A surface of its own: the list is a wall of small text against the page
	 * background otherwise, with nothing to say where it starts and stops.
	 */
	.odds {
		padding: 12px;
		border: 1px solid var(--panel-border);
		border-radius: 10px;
		background: var(--panel-bg);
		font-size: 0.75rem;
	}

	h2 {
		margin: 0 0 4px;
		font-size: 0.72rem;
		font-weight: 700;
		letter-spacing: 0.1em;
		text-transform: uppercase;
		color: var(--fg-muted);
	}

	h3 {
		margin: 12px 0 4px;
		font-size: 0.62rem;
		font-weight: 700;
		letter-spacing: 0.09em;
		text-transform: uppercase;
		color: var(--fg-faint);
	}

	.lede {
		margin: 0;
		font-size: 0.66rem;
		line-height: 1.45;
		color: var(--fg-faint);
	}

	ul {
		list-style: none;
		margin: 0;
		padding: 0;
	}

	li {
		display: grid;
		/* Fixed columns so the bars line up into one readable chart rather
		   than jittering with the width of each percentage. */
		grid-template-columns: 0.9rem 1fr 2.1rem 1.7rem;
		align-items: center;
		gap: 5px;
		padding: 1px 0;
		font-variant-numeric: tabular-nums;
	}

	.letter {
		font-weight: 700;
		text-transform: uppercase;
		letter-spacing: 0.04em;
	}

	.track {
		height: 5px;
		border-radius: 3px;
		background: var(--border);
		overflow: hidden;
	}

	.fill {
		display: block;
		height: 100%;
		border-radius: 3px;
		background: var(--color-correct);
		transition: width 200ms ease;
	}

	.pct {
		text-align: right;
		font-size: 0.68rem;
	}

	.spot {
		text-align: right;
		font-size: 0.66rem;
		color: var(--fg-muted);
	}

	/* Dimmed when the letter is spread across the row: it has a most likely
	   slot, but not one worth betting a guess on. */
	.spot.vague {
		color: var(--fg-faint);
		font-style: italic;
	}

	.out {
		opacity: 0.35;
	}

	.out .fill {
		background: var(--color-absent);
	}
</style>
