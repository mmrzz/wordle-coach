import { ApiError, isAbort } from "../api/client";
import { fetchWords, rate, suggest, type SolverOptions } from "../api/solver";
import {
	BOUNCE_MS,
	BOUNCE_STAGGER_MS,
	FLIP_MS,
	POP_MS,
	ROW_REVEAL_MS,
	TOAST_MS,
	duration,
} from "./anim";
import { seedColors } from "./deduce";
import { keyColors } from "./keyboard";
import { DEFAULT_BETA } from "./temperature";
import {
	MAX_TURNS,
	WINNING_PATTERN,
	WORD_LENGTH,
	nextColor,
	toPattern,
	type Color,
	type Mode,
	type Suggestion,
	type Turn,
} from "./types";

/** How many suggestions the question-mark panel lists. */
export const SUGGESTION_COUNT = 5;

/** How long typing must settle before an unlisted word is sent off to be scored. */
const RATE_DEBOUNCE_MS = 350;

const STORAGE_KEY = "wordle-coach/game/v1";

export type Phase =
	| "loading"
	/** The active row is open for typing, prefilled with our suggestion. */
	| "typing"
	/** The word is locked and the player is relaying the colours back. */
	| "coloring"
	| "won"
	| "lost"
	/** The API could not be reached at all, so nothing else can proceed. */
	| "offline";

/** How a finished game was won, which is only ever a matter of wording. */
export type Outcome =
	/** The player played the word and it came back all green. */
	| "guessed"
	/** Only one answer survived, so the word is known without playing it. */
	| "deduced"
	/** The player pressed the tick to say they had it. */
	| "declared";

/** What a row's score badge shows. Rank is absent until a word has been graded. */
export type Score = {
	bits: number;
	rank?: number;
	poolSize?: number;
	percentile?: number;
};

export type Row = {
	letters: string[];
	/** null means the tile has not been turned over yet. */
	colors: (Color | null)[];
	/** True while the row shows our suggestion rather than the player's typing. */
	ghost: boolean;
	score: Score | null;
};

function emptyRow(): Row {
	return {
		letters: Array(WORD_LENGTH).fill(""),
		colors: Array(WORD_LENGTH).fill(null),
		ghost: false,
		score: null,
	};
}

type Saved = {
	history: Turn[];
	scores: (Score | null)[];
	beta: number;
	mode: Mode;
};

/**
 * The whole coach, as one state machine.
 *
 * Nothing about the position is stored beyond the history: the candidate set,
 * the keyboard colours and every score are recomputed from it, matching how the
 * engine works. That is what makes undo a one-line operation and what lets a
 * refresh pick a game back up from localStorage.
 */
export class Game {
	history = $state<Turn[]>([]);
	rows = $state<Row[]>(Array.from({ length: MAX_TURNS }, emptyRow));
	index = $state(0);
	phase = $state<Phase>("loading");

	/** What the player has typed. Empty means our suggestion is showing. */
	typed = $state("");
	private ghostDismissed = $state(false);

	suggestions = $state<Suggestion[]>([]);
	possibleCount = $state(0);
	remaining = $state<string[]>([]);
	activeScore = $state<Score | null>(null);

	beta = $state(DEFAULT_BETA);
	mode = $state<Mode>("easy");

	thinking = $state(false);
	toast = $state("");
	fatal = $state("");
	panelOpen = $state(false);

	outcome = $state<Outcome | null>(null);
	/** The solution, once it is known. Never guessed at before then. */
	answer = $state<string | null>(null);

	/** Bumped to replay the shake; the value itself carries no meaning. */
	shakeNonce = $state(0);
	/** The row mid-reveal, so its tiles stagger instead of turning at once. */
	revealingRow = $state<number | null>(null);
	/** The row playing the winning bounce. */
	bouncingRow = $state<number | null>(null);

	private allowed = new Set<string>();
	private suggestCtl: AbortController | null = null;
	private rateCtl: AbortController | null = null;
	private rateTimer: ReturnType<typeof setTimeout> | undefined;
	private toastTimer: ReturnType<typeof setTimeout> | undefined;

	get options(): SolverOptions {
		return { mode: this.mode, beta: this.beta };
	}

	/** Our current best guess, shown greyed in the active row until overtyped. */
	get ghostWord(): string {
		return this.suggestions[0]?.word ?? "";
	}

	get showingGhost(): boolean {
		return this.typed === "" && !this.ghostDismissed && this.ghostWord !== "";
	}

	/** The word Enter would submit, whether we suggested it or the player typed it. */
	get activeWord(): string {
		if (this.typed !== "") return this.typed;
		return this.showingGhost ? this.ghostWord : "";
	}

	get over(): boolean {
		return this.phase === "won" || this.phase === "lost";
	}

	get keyColors(): Record<string, Color> {
		return keyColors(this.history);
	}

	/** True once the word list has arrived and the board can be used. */
	get ready(): boolean {
		return this.phase !== "loading" && this.phase !== "offline";
	}

	/**
	 * What to draw in row i. The active row is rendered from activeWord rather
	 * than from stored letters, so the suggestion and the player's typing are
	 * the same code path and cannot drift apart.
	 */
	viewRow(i: number): Row {
		if (i !== this.index || this.phase !== "typing") return this.rows[i];

		const word = this.activeWord;
		return {
			letters: Array.from({ length: WORD_LENGTH }, (_, k) => word[k] ?? ""),
			colors: Array(WORD_LENGTH).fill(null),
			ghost: this.showingGhost,
			score: this.activeScore,
		};
	}

	// -- lifecycle ---------------------------------------------------------

	async init() {
		try {
			const words = await fetchWords();
			this.allowed = new Set(words.allowed);
		} catch (err) {
			this.phase = "offline";
			this.fatal = describe(err);
			return;
		}

		this.restore();
		this.phase = this.history.length >= MAX_TURNS ? "lost" : "typing";
		await this.refresh();
	}

	/** Clears the board but keeps the solver settings, which are a preference. */
	reset() {
		this.cancelPending();
		this.history = [];
		this.rows = Array.from({ length: MAX_TURNS }, emptyRow);
		this.index = 0;
		this.typed = "";
		this.ghostDismissed = false;
		this.activeScore = null;
		this.suggestions = [];
		this.remaining = [];
		this.bouncingRow = null;
		this.revealingRow = null;
		this.outcome = null;
		this.answer = null;
		this.phase = "typing";
		this.persist();
		void this.refresh();
	}

	// -- input -------------------------------------------------------------

	type(letter: string) {
		if (this.phase !== "typing" || this.typed.length >= WORD_LENGTH) return;
		this.typed += letter.toLowerCase();
		this.scheduleRate();
	}

	backspace() {
		// While colouring, backspace reopens the word for editing. It is the
		// only way out of a word submitted by mistake.
		if (this.phase === "coloring") return this.unlock();
		if (this.phase !== "typing") return;

		if (this.typed === "") {
			// Nothing typed, so the only thing to clear is our suggestion.
			if (this.showingGhost) this.ghostDismissed = true;
			return;
		}

		this.typed = this.typed.slice(0, -1);
		this.scheduleRate();
	}

	enter() {
		if (this.phase === "typing") return this.lock();
		if (this.phase === "coloring") return void this.confirm();
	}

	/** Fills the active row from the suggestion panel. */
	apply(word: string) {
		if (this.phase !== "typing") return;
		this.typed = word;
		this.ghostDismissed = false;
		this.panelOpen = false;
		this.scheduleRate();
	}

	/** Advances one tile grey to yellow to green and round again. */
	cycleColor(position: number) {
		if (this.phase !== "coloring") return;
		const row = this.rows[this.index];
		row.colors[position] = nextColor(row.colors[position] ?? "absent");
	}

	/** Steps back one turn, from colouring to typing or from a finished row. */
	undo() {
		if (this.phase === "coloring") return this.unlock();
		if (this.history.length === 0) return;

		this.cancelPending();
		this.history = this.history.slice(0, -1);
		this.index = this.history.length;
		for (let i = this.index + 1; i < MAX_TURNS; i++) this.rows[i] = emptyRow();

		this.rows[this.index].score = null;
		this.typed = "";
		this.ghostDismissed = false;
		this.activeScore = null;
		this.bouncingRow = null;
		this.outcome = null;
		this.answer = null;
		this.phase = "coloring";

		this.persist();
		void this.refresh();
	}

	// -- settings ----------------------------------------------------------

	async setBeta(beta: number) {
		if (beta === this.beta) return;
		this.beta = beta;
		this.persist();
		await this.rescore();
	}

	async setMode(mode: Mode) {
		if (mode === this.mode) return;
		this.mode = mode;
		this.persist();
		await this.rescore();
	}

	/**
	 * Re-runs everything the slider affects. Past rows are graded again because
	 * a score means "how good was this under the current setting", so leaving
	 * old numbers on screen beside new ones would be comparing two rulers.
	 */
	private async rescore() {
		await this.refresh();

		await Promise.all(
			this.history.map((turn, i) => this.grade(i, this.history.slice(0, i), turn.guess)),
		);

		if (this.phase === "typing") this.scheduleRate();
	}

	// -- transitions -------------------------------------------------------

	private lock() {
		const word = this.activeWord;
		if (word.length < WORD_LENGTH) return this.reject("Not enough letters");
		if (!this.allowed.has(word)) return this.reject("Not in word list");

		const row = this.rows[this.index];
		row.letters = word.split("");
		row.ghost = false;
		row.score = this.activeScore;
		// Colours the earlier turns already proved are filled in, so only the
		// genuinely new information has to be entered.
		row.colors = seedColors(this.history, word);

		this.phase = "coloring";
		this.startReveal(this.index);
	}

	/**
	 * The tick: this guess was the answer. Fills the row green and confirms it,
	 * so the winning turn is recorded the same way as one coloured by hand.
	 */
	declareSolved() {
		if (this.phase !== "coloring") return;

		this.rows[this.index].colors = Array(WORD_LENGTH).fill("correct");
		// Let the row finish turning green before the win lands on top of it.
		setTimeout(() => {
			if (this.phase !== "coloring") return;
			this.outcome = "declared";
			void this.confirm();
		}, duration(FLIP_MS));
	}

	/** Returns a locked row to typing, discarding the colours set so far. */
	private unlock() {
		const row = this.rows[this.index];
		this.typed = row.letters.join("").trim();
		row.colors = Array(WORD_LENGTH).fill(null);
		this.phase = "typing";
		this.scheduleRate();
	}

	private async confirm() {
		const row = this.rows[this.index];
		const guess = row.letters.join("");
		const pattern = toPattern(row.colors as Color[]);

		const before = this.history;
		this.history = [...before, { guess, pattern }];
		this.persist();

		// Graded against the position it was played from, so the history passed
		// here is the one that existed before this turn.
		void this.grade(this.index, before, guess);

		if (pattern === WINNING_PATTERN) {
			this.answer = guess;
			this.outcome ??= "guessed";
			this.phase = "won";
			this.celebrate(this.index);
			this.showToast(praise(this.index + 1));
			return;
		}

		if (this.index === MAX_TURNS - 1) {
			this.phase = "lost";
			await this.refresh();
			return;
		}

		this.index++;
		this.typed = "";
		this.ghostDismissed = false;
		this.activeScore = null;
		this.phase = "typing";
		await this.refresh();
	}

	/**
	 * Undoes the last confirmation and reopens its colours. Called when the
	 * server reports that no answer fits, which in practice always means a tile
	 * was set to the wrong colour.
	 */
	private rewind(message: string) {
		if (this.history.length === 0) return;

		this.history = this.history.slice(0, -1);
		this.index = this.history.length;
		this.typed = "";
		this.ghostDismissed = false;
		this.bouncingRow = null;
		this.phase = "coloring";
		this.persist();
		this.showToast(message);
	}

	// -- server work -------------------------------------------------------

	/** Fetches the ranking for the current position. */
	async refresh() {
		this.suggestCtl?.abort();
		const ctl = new AbortController();
		this.suggestCtl = ctl;
		this.thinking = true;

		try {
			const res = await suggest(this.history, this.options, SUGGESTION_COUNT, ctl.signal);
			this.suggestions = res.suggestions;
			this.possibleCount = res.possibleCount;
			this.remaining = res.remaining ?? [];

			/*
			 * One survivor means the answer is settled, whether or not it has
			 * been played, so there is nothing left to coach and the game is
			 * won. Restricted to the typing phase so that stepping back through
			 * finished turns does not keep re-winning the game.
			 */
			if (this.phase === "typing" && this.possibleCount === 1 && this.remaining.length === 1) {
				this.answer = this.remaining[0];
				this.outcome = "deduced";
				// Set before filling the row: while the phase is still typing,
				// the active row is drawn from activeWord and would ignore it.
				this.phase = "won";
				this.revealAnswer(this.answer);
				return;
			}

			if (this.phase === "typing") this.scheduleRate();
		} catch (err) {
			if (isAbort(err)) return;
			if (err instanceof ApiError && err.code === "inconsistent_history") {
				this.rewind("Those colours can't all be right — check the tiles");
				return;
			}
			this.showToast(describe(err));
		} finally {
			if (this.suggestCtl === ctl) this.thinking = false;
		}
	}

	/** Grades one played word and hangs the result on its row. */
	private async grade(row: number, before: Turn[], guess: string) {
		try {
			const res = await rate(before, guess, this.options);
			this.rows[row].score = {
				bits: res.playedBits,
				rank: res.playedRank,
				poolSize: res.poolSize,
				percentile: res.percentile,
			};
		} catch (err) {
			// A missing score is cosmetic, so a failure here stays quiet rather
			// than interrupting a game that is otherwise fine.
			if (!isAbort(err)) console.warn("rate failed", err);
		}
	}

	/**
	 * Scores whatever is in the active row. Words already in the suggestion
	 * list are scored for free from the response we hold, so only a word the
	 * player typed themselves costs a request, and only once typing settles.
	 */
	private scheduleRate() {
		clearTimeout(this.rateTimer);
		this.rateCtl?.abort();
		this.activeScore = null;

		const word = this.activeWord;
		if (word.length !== WORD_LENGTH || !this.allowed.has(word)) return;

		const known = this.suggestions.findIndex((s) => s.word === word);
		if (known !== -1) {
			// The suggestion list is the top of the fully sorted pool, so its
			// position is the word's rank.
			this.activeScore = { bits: this.suggestions[known].bits, rank: known + 1 };
			return;
		}

		this.rateTimer = setTimeout(() => void this.rateActive(word), RATE_DEBOUNCE_MS);
	}

	private async rateActive(word: string) {
		const ctl = new AbortController();
		this.rateCtl = ctl;

		try {
			const res = await rate(this.history, word, this.options, ctl.signal);
			// Typing may have moved on while this was in flight.
			if (this.activeWord !== word) return;
			this.activeScore = {
				bits: res.playedBits,
				rank: res.playedRank,
				poolSize: res.poolSize,
				percentile: res.percentile,
			};
		} catch (err) {
			if (!isAbort(err)) console.warn("rate failed", err);
		}
	}

	// -- effects -----------------------------------------------------------

	private reject(message: string) {
		this.shakeNonce++;
		this.showToast(message);
	}

	/**
	 * Puts the one surviving answer on the board rather than only naming it.
	 *
	 * The letters land first, then turn over green left to right and take the
	 * winning bounce, so the word arrives the same way a played one would. It
	 * is display only and never joins the history, because the player did not
	 * play it: undo clears the row along with the turn that produced it.
	 */
	private revealAnswer(word: string) {
		const target = this.index;
		const row = this.rows[target];

		row.letters = word.split("");
		row.colors = Array(WORD_LENGTH).fill(null);
		row.ghost = false;
		row.score = null;

		setTimeout(() => {
			if (this.phase !== "won" || this.index !== target) return;

			this.rows[target].colors = Array(WORD_LENGTH).fill("correct");
			this.startReveal(target);

			setTimeout(() => {
				if (this.phase !== "won" || this.index !== target) return;
				this.celebrate(target);
				this.showToast(`It had to be ${word.toUpperCase()}`);
			}, duration(ROW_REVEAL_MS));
		}, duration(POP_MS) + 80);
	}

	private startReveal(row: number) {
		this.revealingRow = row;
		setTimeout(() => {
			if (this.revealingRow === row) this.revealingRow = null;
		}, duration(ROW_REVEAL_MS) + 50);
	}

	private celebrate(row: number) {
		this.bouncingRow = row;
		setTimeout(
			() => {
				if (this.bouncingRow === row) this.bouncingRow = null;
			},
			duration(BOUNCE_MS + BOUNCE_STAGGER_MS * WORD_LENGTH) + 50,
		);
	}

	showToast(message: string) {
		clearTimeout(this.toastTimer);
		this.toast = message;
		this.toastTimer = setTimeout(() => (this.toast = ""), TOAST_MS);
	}

	private cancelPending() {
		clearTimeout(this.rateTimer);
		this.rateCtl?.abort();
		this.suggestCtl?.abort();
	}

	// -- persistence -------------------------------------------------------

	private persist() {
		const saved: Saved = {
			history: this.history,
			scores: this.rows.map((r) => r.score),
			beta: this.beta,
			mode: this.mode,
		};
		try {
			localStorage.setItem(STORAGE_KEY, JSON.stringify(saved));
		} catch {
			// Private browsing and full quotas both throw here. Losing the save
			// is not worth breaking the game over.
		}
	}

	private restore() {
		let saved: Saved;
		try {
			const raw = localStorage.getItem(STORAGE_KEY);
			if (!raw) return;
			saved = JSON.parse(raw) as Saved;
		} catch {
			return;
		}

		if (!Array.isArray(saved.history)) return;
		// A saved game from a different word list would filter to nothing, so
		// anything unrecognised is dropped rather than trusted.
		const history = saved.history.filter(
			(t) =>
				typeof t?.guess === "string" &&
				this.allowed.has(t.guess) &&
				typeof t?.pattern === "string" &&
				/^[gyb]{5}$/.test(t.pattern),
		);
		if (history.length !== saved.history.length) return;

		this.history = history;
		this.beta = typeof saved.beta === "number" ? saved.beta : DEFAULT_BETA;
		this.mode = saved.mode === "hard" ? "hard" : "easy";

		history.forEach((turn, i) => {
			this.rows[i] = {
				letters: turn.guess.split(""),
				colors: turn.pattern.split("").map(toColor),
				ghost: false,
				score: saved.scores?.[i] ?? null,
			};
		});

		this.index = Math.min(history.length, MAX_TURNS - 1);
		if (history.at(-1)?.pattern === WINNING_PATTERN) {
			this.phase = "won";
			this.outcome = "guessed";
			this.answer = history.at(-1)!.guess;
		}
	}
}

function toColor(code: string): Color {
	if (code === "g") return "correct";
	if (code === "y") return "present";
	return "absent";
}

function praise(turns: number): string {
	return ["Genius", "Magnificent", "Impressive", "Splendid", "Great", "Phew"][turns - 1] ?? "Solved";
}

function describe(err: unknown): string {
	if (err instanceof ApiError && err.code === "network") {
		return "Can't reach the solver — is the Go server running?";
	}
	return err instanceof Error ? err.message : "Something went wrong";
}
