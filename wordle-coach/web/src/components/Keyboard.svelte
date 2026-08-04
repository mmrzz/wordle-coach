<script lang="ts">
	import { KEY_ROWS } from "../lib/keyboard";
	import type { Color } from "../lib/types";

	type Props = {
		colors: Record<string, Color>;
		/** Letters are dead while the player is relaying colours back. */
		lettersDisabled?: boolean;
		onKey: (value: string) => void;
	};

	let { colors, lettersDisabled = false, onKey }: Props = $props();

	function isAction(value: string): boolean {
		return value === "Enter" || value === "Backspace";
	}
</script>

<div class="keyboard" role="group" aria-label="Keyboard">
	{#each KEY_ROWS as row, r (r)}
		<div class="krow">
			{#if r === 1}<div class="spacer"></div>{/if}
			{#each row as key (key.value)}
				<button
					class="key"
					class:wide={key.wide}
					data-state={colors[key.value] ?? ""}
					disabled={lettersDisabled && !isAction(key.value)}
					onclick={() => onKey(key.value)}
					aria-label={key.value === "Backspace" ? "Backspace" : key.value}
				>
					{key.label}
				</button>
			{/each}
			{#if r === 1}<div class="spacer"></div>{/if}
		</div>
	{/each}
</div>

<style>
	.keyboard {
		display: flex;
		flex-direction: column;
		gap: 8px;
		width: 100%;
		max-width: 500px;
		margin: 0 auto;
		padding: 0 8px 8px;
		height: var(--keyboard-height);
		touch-action: manipulation;
	}

	.krow {
		display: flex;
		gap: 6px;
		flex: 1;
	}

	/* Half-key indents that centre the middle row, as in the real layout. */
	.spacer {
		flex: 0.5;
	}

	.key {
		display: flex;
		align-items: center;
		justify-content: center;
		flex: 1;
		min-width: 0;
		height: 100%;
		border-radius: 4px;
		background: var(--key-bg);
		color: var(--key-fg);
		font-size: 0.8rem;
		font-weight: 700;
		text-transform: uppercase;
		user-select: none;
		transition: background 40ms ease;
	}

	.key.wide {
		flex: 1.5;
		font-size: 0.72rem;
	}

	.key:hover:not(:disabled) {
		filter: brightness(1.08);
	}

	.key:active:not(:disabled) {
		filter: brightness(0.92);
	}

	.key:disabled {
		opacity: 0.45;
		cursor: default;
	}

	.key[data-state="absent"] {
		background: var(--color-absent);
		color: #fff;
	}

	.key[data-state="present"] {
		background: var(--color-present);
		color: #fff;
	}

	.key[data-state="correct"] {
		background: var(--color-correct);
		color: #fff;
	}

	@media (max-width: 400px) {
		.keyboard {
			gap: 6px;
			padding: 0 4px 6px;
		}
		.krow {
			gap: 4px;
		}
		.key {
			font-size: 0.7rem;
		}
	}
</style>
