<script lang="ts">
	type Props = { message: string };
	let { message }: Props = $props();
</script>

<div class="host" aria-live="polite" aria-atomic="true">
	{#if message}
		<div class="toast">{message}</div>
	{/if}
</div>

<style>
	.host {
		position: fixed;
		top: calc(var(--header-height) + 12px);
		left: 0;
		right: 0;
		z-index: 70;
		display: flex;
		justify-content: center;
		pointer-events: none;
	}

	.toast {
		max-width: min(90vw, 340px);
		padding: 12px 16px;
		border-radius: 4px;
		background: var(--toast-bg);
		color: var(--toast-fg);
		font-size: 0.85rem;
		font-weight: 700;
		text-align: center;
		animation: FadeUp 120ms ease-out;
	}
</style>
