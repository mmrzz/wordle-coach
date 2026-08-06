<script lang="ts">
	import { onMount } from "svelte";
	import Board from "./components/Board.svelte";
	import Header from "./components/Header.svelte";
	import HelpModal from "./components/HelpModal.svelte";
	import Keyboard from "./components/Keyboard.svelte";
	import LetterOdds from "./components/LetterOdds.svelte";
	import OffBookPrompt from "./components/OffBookPrompt.svelte";
	import Shortlist from "./components/Shortlist.svelte";
	import StatusBar from "./components/StatusBar.svelte";
	import SuggestionPanel from "./components/SuggestionPanel.svelte";
	import TemperatureDial from "./components/TemperatureDial.svelte";
	import Toast from "./components/Toast.svelte";
	import WildOpener from "./components/WildOpener.svelte";
	import { Game } from "./lib/game.svelte";

	type Theme = "light" | "dark";
	const THEME_KEY = "wordle-coach/theme";

	const game = new Game();
	let helpOpen = $state(false);

	// Unset means follow the system, which is what the stylesheet does when no
	// data-theme attribute is present.
	let theme = $state<Theme | null>(null);
	let systemTheme = $state<Theme>("light");

	onMount(() => {
		void game.init();

		const saved = localStorage.getItem(THEME_KEY);
		if (saved === "light" || saved === "dark") theme = saved;

		const dark = matchMedia("(prefers-color-scheme: dark)");
		systemTheme = dark.matches ? "dark" : "light";

		const follow = (e: MediaQueryListEvent) => (systemTheme = e.matches ? "dark" : "light");
		dark.addEventListener("change", follow);
		return () => dark.removeEventListener("change", follow);
	});

	$effect(() => {
		if (theme) document.documentElement.dataset.theme = theme;
		else delete document.documentElement.dataset.theme;
	});

	/** What is actually on screen, whether chosen here or inherited. */
	const activeTheme = $derived(theme ?? systemTheme);

	function toggleTheme() {
		theme = activeTheme === "dark" ? "light" : "dark";
		localStorage.setItem(THEME_KEY, theme);
	}

	/** Routes both the physical keyboard and the on-screen one through one path. */
	function press(value: string) {
		if (!game.ready || game.over) return;

		if (value === "Enter") return game.enter();
		if (value === "Backspace") return game.backspace();
		if (/^[a-zA-Z]$/.test(value)) return game.type(value);
	}

	function onkeydown(event: KeyboardEvent) {
		// Let dialogs and the dial handle their own keys, and leave browser
		// shortcuts alone.
		if (game.panelOpen || helpOpen || game.offBookPrompt) return;
		if (event.metaKey || event.ctrlKey || event.altKey) return;

		const target = event.target as HTMLElement | null;
		if (target?.tagName === "INPUT") return;

		if (event.key === "Enter" || event.key === "Backspace" || /^[a-zA-Z]$/.test(event.key)) {
			event.preventDefault();
			press(event.key);
		}
	}

</script>

<svelte:window {onkeydown} />

<div class="app">
	<Header
		theme={activeTheme}
		canUndo={game.history.length > 0 || game.phase === "coloring"}
		onHelp={() => (helpOpen = true)}
		onUndo={() => game.undo()}
		onReset={() => game.reset()}
		onTheme={toggleTheme}
	/>

	<StatusBar {game} />

	<div class="stage">
		<WildOpener {game} />
		<Board {game} />
	</div>

	<Keyboard
		colors={game.keyColors}
		lettersDisabled={game.phase !== "typing"}
		onKey={press}
	/>
</div>

<div class="rail left">
	<Shortlist {game} />
</div>

<div class="rail right">
	<LetterOdds {game} />
</div>

<Toast message={game.toast} />
<SuggestionPanel {game} />
<OffBookPrompt {game} />
<HelpModal open={helpOpen} onclose={() => (helpOpen = false)} />
<TemperatureDial {game} />

<style>
	/*
	 * Wider than the real game's shell to leave room for the score column
	 * either side of the board; the board itself keeps its original size.
	 */
	.app {
		display: flex;
		flex-direction: column;
		height: 100%;
		max-width: 560px;
		margin: 0 auto;
	}

	/*
	 * The wild opener belongs to the board, not to the status line above it, so
	 * the two are centred together as one block. Without this the opener would
	 * be pinned under the answers-left count while the board floated free in the
	 * space below, which read as if the offer were part of the status line.
	 */
	.stage {
		display: flex;
		flex-direction: column;
		justify-content: center;
		flex: 1;
		min-height: 0;
	}

	/*
	 * The rails are fixed rather than columns of a flex row, so the board sits
	 * exactly where the real game puts it whether or not the shortlist has
	 * anything to say. They need the shell's 560px plus their own width twice
	 * over; below that there is no room and they are hidden, with the ? panel
	 * carrying the same information instead.
	 */
	.rail {
		position: fixed;
		top: calc(var(--header-height) + 12px);
		/* Clear of the temperature dial, which is fixed to the bottom left. */
		bottom: 76px;
		width: var(--rail-width);
		overflow-y: auto;
		overscroll-behavior: contain;
		z-index: 20;
	}

	.left {
		left: 16px;
	}

	.right {
		right: 32px;
	}

	@media (max-width: 1060px) {
		.rail {
			display: none;
		}
	}
</style>
