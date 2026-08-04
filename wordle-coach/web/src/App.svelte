<script lang="ts">
	import { onMount } from "svelte";
	import Board from "./components/Board.svelte";
	import Header from "./components/Header.svelte";
	import HelpModal from "./components/HelpModal.svelte";
	import Keyboard from "./components/Keyboard.svelte";
	import StatusBar from "./components/StatusBar.svelte";
	import SuggestionPanel from "./components/SuggestionPanel.svelte";
	import TemperatureDial from "./components/TemperatureDial.svelte";
	import Toast from "./components/Toast.svelte";
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
		if (game.panelOpen || helpOpen) return;
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
	<Board {game} />

	<Keyboard
		colors={game.keyColors}
		lettersDisabled={game.phase !== "typing"}
		onKey={press}
	/>
</div>

<Toast message={game.toast} />
<SuggestionPanel {game} />
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
</style>
