import { getJSON, postJSON } from "./client";
import type {
	Mode,
	RateResponse,
	SuggestResponse,
	Turn,
	WordsResponse,
} from "../lib/types";

/** Settings that apply to every scoring request. */
export type SolverOptions = {
	mode: Mode;
	/** The temperature slider's Renyi order. */
	beta: number;
};

/** Asks for the best next guesses given every turn played so far. */
export function suggest(
	history: Turn[],
	opts: SolverOptions,
	limit: number,
	signal?: AbortSignal,
): Promise<SuggestResponse> {
	return postJSON<SuggestResponse>(
		"/api/suggest",
		{ mode: opts.mode, beta: opts.beta, history, limit },
		signal,
	);
}

/**
 * Grades one word from the position it was played, so history must hold the
 * turns before it and not the turn itself.
 */
export function rate(
	history: Turn[],
	played: string,
	opts: SolverOptions,
	signal?: AbortSignal,
): Promise<RateResponse> {
	return postJSON<RateResponse>(
		"/api/rate",
		{ mode: opts.mode, beta: opts.beta, history, played },
		signal,
	);
}

/**
 * Fetches the guess list once so typing can be validated without a round trip.
 * It is what makes an unknown word shake immediately, the way the real game
 * does.
 */
export function fetchWords(signal?: AbortSignal): Promise<WordsResponse> {
	return getJSON<WordsResponse>("/api/words", signal);
}
