<script lang="ts">
	import type { Snippet } from "svelte";

	type Props = {
		open: boolean;
		title: string;
		onclose: () => void;
		children: Snippet;
	};

	let { open, title, onclose, children }: Props = $props();

	let panel = $state<HTMLDivElement | null>(null);

	// Focus moves into the dialog when it opens, so keyboard and screen-reader
	// users land inside it rather than being left behind on the board.
	$effect(() => {
		if (open && panel) panel.focus();
	});

	function onkeydown(event: KeyboardEvent) {
		if (event.key === "Escape") {
			event.stopPropagation();
			onclose();
		}
	}
</script>

{#if open}
	<div
		class="backdrop"
		role="presentation"
		onclick={(event) => {
			// Only a click on the backdrop itself closes; one that bubbled up
			// from inside the dialog is a click on the dialog.
			if (event.target === event.currentTarget) onclose();
		}}
		onkeydown={onkeydown}
	>
		<div
			class="panel"
			role="dialog"
			aria-modal="true"
			aria-label={title}
			tabindex="-1"
			bind:this={panel}
		>
			<header>
				<h2>{title}</h2>
				<button class="close" onclick={onclose} aria-label="Close">×</button>
			</header>
			<div class="body">
				{@render children()}
			</div>
		</div>
	</div>
{/if}

<style>
	.backdrop {
		position: fixed;
		inset: 0;
		z-index: 60;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 16px;
		background: var(--overlay);
	}

	.panel {
		width: 100%;
		max-width: 420px;
		max-height: 85vh;
		overflow-y: auto;
		background: var(--bg-raised);
		color: var(--fg);
		border: 1px solid var(--border);
		border-radius: 8px;
		box-shadow: 0 4px 30px var(--shadow);
		animation: SlideUp 160ms ease-out;
		/*
		 * Focused on open so the keyboard lands inside, but the dialog is not
		 * itself a control, so it should not wear a focus ring. The things
		 * inside it still get theirs.
		 */
		outline: none;
	}

	header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 14px 16px 8px;
	}

	h2 {
		margin: 0;
		font-size: 0.95rem;
		font-weight: 700;
		letter-spacing: 0.08em;
		text-transform: uppercase;
	}

	.close {
		font-size: 1.6rem;
		line-height: 1;
		color: var(--fg-muted);
		padding: 0 4px;
	}

	.close:hover {
		color: var(--fg);
	}

	.body {
		padding: 0 16px 18px;
	}
</style>
