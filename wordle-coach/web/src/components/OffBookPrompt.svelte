<script lang="ts">
	import type { Game } from "../lib/game.svelte";
	import Modal from "./Modal.svelte";

	type Props = { game: Game };
	let { game }: Props = $props();

	/*
	 * The last word played, which is the one whose colours are in question:
	 * whatever went wrong, it went wrong on the turn that has just been
	 * confirmed. Naming it saves the player hunting for the row.
	 */
	const suspect = $derived(game.history.at(-1)?.guess.toUpperCase() ?? "");

	// Closing the dialog any other way is the cautious answer, since it is the
	// one that changes nothing about how the game is being solved.
	const check = () => game.checkTheColours();
</script>

<Modal open={game.offBookPrompt} title="Check the colours" onclose={check}>
	<p>
		No official answer fits the colours you have given{suspect
			? ` for ${suspect}`
			: ""}. That nearly always means a tile is the wrong colour, so it is worth
		another look first.
	</p>

	<p class="aside">
		If you are sure they are right, the game you are playing is not drawing its
		answer from the official list of 2,315 — plenty of clones use the far larger
		guess list instead. We can solve against that instead.
	</p>

	<div class="choices">
		<button class="primary" onclick={check}>Let me look again</button>
		<button class="secondary" onclick={() => void game.goOffBook()}>
			I'm sure — go off the books
		</button>
	</div>

	<p class="note">
		Going off the books widens the answer set to every legal word, so expect
		more guesses to be needed. A new game starts back on the official list.
	</p>
</Modal>

<style>
	p {
		margin: 0 0 12px;
		font-size: 0.86rem;
		line-height: 1.55;
		color: var(--fg);
	}

	.aside {
		padding: 8px 10px;
		border-radius: 6px;
		background: color-mix(in srgb, var(--fg) 6%, transparent);
		font-size: 0.82rem;
		color: var(--fg-muted);
	}

	/*
	 * Stacked, with looking again on top and wearing the filled button: the two
	 * answers are not equal weight. A mistyped colour is an everyday slip and an
	 * unofficial answer is not, so the cautious reply is the one offered first
	 * and the wider list has to be chosen deliberately.
	 */
	.choices {
		display: flex;
		flex-direction: column;
		gap: 8px;
		margin: 16px 0 12px;
	}

	.choices button {
		padding: 10px 12px;
		border-radius: 6px;
		font-size: 0.84rem;
		font-weight: 700;
		letter-spacing: 0.02em;
	}

	.primary {
		background: var(--color-correct);
		color: #ffffff;
	}

	.secondary {
		border: 1px solid var(--border-strong);
		color: var(--fg);
	}

	.choices button:hover {
		filter: brightness(1.08);
	}

	.note {
		margin: 0;
		font-size: 0.74rem;
		line-height: 1.5;
		color: var(--fg-faint);
	}
</style>
