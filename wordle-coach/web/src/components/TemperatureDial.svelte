<script lang="ts">
	import type { Game } from "../lib/game.svelte";
	import {
		LOG_STEP,
		MAX_LOG,
		MIN_LOG,
		TICKS,
		band,
		fromSlider,
		toSlider,
	} from "../lib/temperature";

	type Props = { game: Game };
	let { game }: Props = $props();

	let open = $state(false);
	let popover = $state<HTMLDivElement | null>(null);

	const position = $derived(toSlider(game.beta));
	const current = $derived(band(game.beta));

	function onInput(event: Event) {
		const value = Number((event.currentTarget as HTMLInputElement).value);
		void game.setBeta(fromSlider(value));
	}

	// Clicking anywhere else puts the dial away, the way a popover should
	// behave. Registered only while it is open.
	$effect(() => {
		if (!open) return;

		const dismiss = (event: MouseEvent) => {
			if (popover && !popover.contains(event.target as Node)) open = false;
		};
		const escape = (event: KeyboardEvent) => {
			if (event.key === "Escape") open = false;
		};

		document.addEventListener("mousedown", dismiss);
		document.addEventListener("keydown", escape);
		return () => {
			document.removeEventListener("mousedown", dismiss);
			document.removeEventListener("keydown", escape);
		};
	});
</script>

<div class="dial" bind:this={popover}>
	{#if open}
		<div class="popover" role="group" aria-label="Solver temperature">
			<div class="head">
				<span class="name">{current.name}</span>
				<span class="beta">β {game.beta}</span>
			</div>

			<input
				type="range"
				min={MIN_LOG}
				max={MAX_LOG}
				step={LOG_STEP}
				value={position}
				oninput={onInput}
				aria-label="Temperature"
				aria-valuetext={`${current.name}, beta ${game.beta}`}
			/>

			<div class="ticks" aria-hidden="true">
				{#each TICKS as tick (tick.label)}
					<span
						style:left={`${((tick.position - MIN_LOG) / (MAX_LOG - MIN_LOG)) * 100}%`}
					>
						{tick.label}
					</span>
				{/each}
			</div>

			<p class="blurb">{current.blurb}</p>

			<!--
				The dial only matters when the two measures disagree, so it is
				worth saying here what each of them is.
			-->
			<dl class="legend">
				<dt>bits</dt>
				<dd>how much a guess tells you — each bit halves the field</dd>
				<dt>leaves</dt>
				<dd>how many answers you expect to be left afterwards</dd>
			</dl>
			<p class="trade">
				Towards <em>Bold</em> this chases bits; towards <em>Cautious</em> it
				protects "leaves".
			</p>

			<div class="modes">
				<span class="label">Guess pool</span>
				<div class="segmented" role="group" aria-label="Guess pool">
					<button
						class:on={game.mode === "easy"}
						onclick={() => game.setMode("easy")}
						title="Any word may be guessed"
					>
						Easy
					</button>
					<button
						class:on={game.mode === "hard"}
						onclick={() => game.setMode("hard")}
						title="Guesses must reuse revealed letters"
					>
						Hard
					</button>
				</div>
			</div>
		</div>
	{/if}

	<button
		class="knob"
		class:open
		onclick={() => (open = !open)}
		aria-expanded={open}
		aria-label={`Temperature: ${current.name}. Adjust how the solver scores.`}
		title={`Temperature — ${current.name}`}
	>
		<svg viewBox="0 0 24 24" width="22" height="22" aria-hidden="true">
			<path
				d="M14 14.76V5a2 2 0 1 0-4 0v9.76a4 4 0 1 0 4 0z"
				fill="none"
				stroke="currentColor"
				stroke-width="1.8"
				stroke-linejoin="round"
			/>
			<circle cx="12" cy="17.5" r="1.9" fill="currentColor" />
		</svg>
	</button>
</div>

<style>
	.dial {
		position: fixed;
		left: 16px;
		bottom: 16px;
		z-index: 50;
	}

	/*
	 * The knob shares the bottom left corner with Enter. It only gets away with
	 * it while the keyboard, capped at 500px and centred, leaves the knob's own
	 * width clear either side of it — below 624px of viewport the two overlap
	 * and the knob takes the presses meant for Enter. Narrower than that it goes
	 * above the keyboard instead, into the slack the stage leaves under the
	 * board, where it is still clear of the board itself: the board is centred
	 * and never comes within the aside width of the edge.
	 */
	@media (max-width: 640px) {
		.dial {
			bottom: calc(var(--keyboard-height) + 8px);
		}
	}

	.knob {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 46px;
		height: 46px;
		border-radius: 50%;
		border: 2px solid var(--border);
		background: var(--bg-raised);
		color: var(--fg-muted);
		box-shadow: 0 2px 10px var(--shadow);
		transition: color 140ms ease, border-color 140ms ease, transform 140ms ease;
	}

	.knob:hover,
	.knob.open {
		color: var(--fg);
		border-color: var(--border-strong);
		transform: scale(1.06);
	}

	.popover {
		position: absolute;
		left: 0;
		bottom: 58px;
		width: 268px;
		padding: 14px;
		border: 1px solid var(--border);
		border-radius: 10px;
		background: var(--bg-raised);
		box-shadow: 0 6px 26px var(--shadow);
		animation: SlideUp 140ms ease-out;
	}

	.head {
		display: flex;
		align-items: baseline;
		justify-content: space-between;
		margin-bottom: 10px;
	}

	.name {
		font-size: 0.9rem;
		font-weight: 700;
		letter-spacing: 0.06em;
		text-transform: uppercase;
	}

	.beta {
		font-size: 0.72rem;
		color: var(--fg-muted);
		font-variant-numeric: tabular-nums;
	}

	input[type="range"] {
		width: 100%;
		margin: 0;
		accent-color: var(--color-correct);
	}

	.ticks {
		position: relative;
		height: 1rem;
		margin-top: 2px;
		font-size: 0.62rem;
		color: var(--fg-faint);
	}

	.ticks span {
		position: absolute;
		transform: translateX(-50%);
		white-space: nowrap;
	}

	/* The end labels would otherwise hang off the popover. */
	.ticks span:first-child {
		transform: none;
	}
	.ticks span:last-child {
		transform: translateX(-100%);
	}

	.blurb {
		margin: 8px 0 0;
		font-size: 0.72rem;
		line-height: 1.5;
		color: var(--fg-muted);
	}

	.legend {
		margin: 10px 0 0;
		padding-top: 10px;
		border-top: 1px solid var(--divider);
		font-size: 0.68rem;
		line-height: 1.45;
	}

	.legend dt {
		font-weight: 700;
		font-variant: small-caps;
		letter-spacing: 0.04em;
		color: var(--fg);
	}

	.legend dt:not(:first-child) {
		margin-top: 5px;
	}

	.legend dd {
		margin: 0;
		color: var(--fg-muted);
	}

	.trade {
		margin: 8px 0 0;
		font-size: 0.68rem;
		line-height: 1.45;
		color: var(--fg-muted);
	}

	.trade em {
		color: var(--fg);
		font-style: normal;
		font-weight: 700;
	}

	.modes {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-top: 12px;
		padding-top: 12px;
		border-top: 1px solid var(--divider);
	}

	.label {
		font-size: 0.72rem;
		color: var(--fg-muted);
	}

	.segmented {
		display: flex;
		border: 1px solid var(--border);
		border-radius: 6px;
		overflow: hidden;
	}

	.segmented button {
		padding: 4px 10px;
		font-size: 0.7rem;
		font-weight: 700;
		color: var(--fg-muted);
	}

	.segmented button.on {
		background: var(--color-correct);
		color: #fff;
	}
</style>
