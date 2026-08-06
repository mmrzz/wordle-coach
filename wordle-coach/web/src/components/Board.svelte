<script lang="ts">
	import type { Game } from "../lib/game.svelte";
	import { MAX_TURNS } from "../lib/types";
	import Row from "./Row.svelte";

	type Props = { game: Game };
	let { game }: Props = $props();

	const rows = $derived(Array.from({ length: MAX_TURNS }, (_, i) => i));
</script>

<div class="board-area">
	<div class="board">
		{#each rows as i (i)}
			<Row
				row={game.viewRow(i)}
				index={i}
				active={i === game.index && !game.over}
				coloring={i === game.index && game.phase === "coloring"}
				staggered={game.revealingRow === i}
				loading={i === game.index && game.awaitingSuggestion}
				bouncing={game.bouncingRow === i}
				shakeNonce={game.shakeNonce}
				pendingScore={i === game.index && game.phase === "typing" && game.thinking}
				onTileClick={(position) => game.cycleColor(position)}
				onAsk={() => (game.panelOpen = true)}
				onSolved={() => game.declareSolved()}
			/>
		{/each}
	</div>
</div>

<style>
	/*
	 * The badges and the ? button hang outside the board so that the grid keeps
	 * the real game's centring. Symmetric padding of exactly their width is
	 * what reserves room for them: the board stays centred in the viewport and
	 * shrinks rather than letting the right-hand column fall off the screen.
	 *
	 * It takes only the height it needs: the centring of the board within the
	 * page is the stage's job, so that whatever else the stage holds stays with
	 * the board rather than being pushed away from it.
	 */
	.board-area {
		display: flex;
		align-items: center;
		justify-content: center;
		flex: 0 1 auto;
		min-height: 0;
		padding: 10px var(--aside-width);
	}

	.board {
		display: grid;
		grid-template-rows: repeat(6, 1fr);
		gap: 5px;
		width: min(350px, 100%);
		max-height: 100%;
		aspect-ratio: 350 / 420;
	}
</style>
