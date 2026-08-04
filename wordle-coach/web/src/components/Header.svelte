<script lang="ts">
	type Theme = "light" | "dark";

	type Props = {
		theme: Theme;
		canUndo: boolean;
		onHelp: () => void;
		onUndo: () => void;
		onReset: () => void;
		onTheme: () => void;
	};

	let { theme, canUndo, onHelp, onUndo, onReset, onTheme }: Props = $props();
</script>

<header>
	<div class="side">
		<button onclick={onHelp} aria-label="How to use this" title="How to use this">
			<svg viewBox="0 0 24 24" width="22" height="22" aria-hidden="true">
				<circle cx="12" cy="12" r="9" fill="none" stroke="currentColor" stroke-width="1.8" />
				<path
					d="M9.5 9.4a2.6 2.6 0 1 1 3.4 2.5c-.6.2-.9.7-.9 1.3v.6"
					fill="none"
					stroke="currentColor"
					stroke-width="1.8"
					stroke-linecap="round"
				/>
				<circle cx="12" cy="16.6" r="1.05" fill="currentColor" />
			</svg>
		</button>
	</div>

	<h1>Wordle Coach</h1>

	<div class="side end">
		<button onclick={onUndo} disabled={!canUndo} aria-label="Undo the last turn" title="Undo">
			<svg viewBox="0 0 24 24" width="22" height="22" aria-hidden="true">
				<path
					d="M4 9h9a5 5 0 0 1 0 10h-3M4 9l4-4M4 9l4 4"
					fill="none"
					stroke="currentColor"
					stroke-width="1.8"
					stroke-linecap="round"
					stroke-linejoin="round"
				/>
			</svg>
		</button>

		<button onclick={onTheme} aria-label="Switch theme" title="Switch theme">
			{#if theme === "dark"}
				<svg viewBox="0 0 24 24" width="22" height="22" aria-hidden="true">
					<circle cx="12" cy="12" r="4.2" fill="none" stroke="currentColor" stroke-width="1.8" />
					<path
						d="M12 2.6v2.2M12 19.2v2.2M2.6 12h2.2M19.2 12h2.2M5.4 5.4l1.6 1.6M17 17l1.6 1.6M18.6 5.4L17 7M7 17l-1.6 1.6"
						stroke="currentColor"
						stroke-width="1.8"
						stroke-linecap="round"
					/>
				</svg>
			{:else}
				<svg viewBox="0 0 24 24" width="22" height="22" aria-hidden="true">
					<path
						d="M20 14.2A8.4 8.4 0 0 1 9.8 4 8.4 8.4 0 1 0 20 14.2z"
						fill="none"
						stroke="currentColor"
						stroke-width="1.8"
						stroke-linejoin="round"
					/>
				</svg>
			{/if}
		</button>

		<button onclick={onReset} aria-label="Start a new game" title="New game">
			<svg viewBox="0 0 24 24" width="22" height="22" aria-hidden="true">
				<path
					d="M20 12a8 8 0 1 1-2.6-5.9M20 4v4.4h-4.4"
					fill="none"
					stroke="currentColor"
					stroke-width="1.8"
					stroke-linecap="round"
					stroke-linejoin="round"
				/>
			</svg>
		</button>
	</div>
</header>

<style>
	header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		height: var(--header-height);
		padding: 0 12px;
		border-bottom: 1px solid var(--divider);
		flex: none;
	}

	h1 {
		margin: 0;
		font-size: clamp(1.1rem, 5vw, 1.65rem);
		font-weight: 700;
		letter-spacing: 0.02em;
		text-align: center;
		white-space: nowrap;
	}

	.side {
		display: flex;
		align-items: center;
		gap: 4px;
		flex: 1;
	}

	.side.end {
		justify-content: flex-end;
	}

	button {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 34px;
		height: 34px;
		border-radius: 4px;
		color: var(--fg);
	}

	button:hover:not(:disabled) {
		background: color-mix(in srgb, var(--fg) 10%, transparent);
	}

	button:disabled {
		opacity: 0.3;
		cursor: default;
	}
</style>
