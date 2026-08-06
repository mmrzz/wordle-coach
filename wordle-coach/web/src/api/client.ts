/*
 * Empty in production, because the page and the API are served from the same
 * origin there and a relative path is what keeps them together however the
 * deployment is named. In development the API is a separate process on another
 * port, which is the only reason this is ever an absolute URL. VITE_API_URL
 * overrides both, for pointing a local page at a deployed server.
 */
export const BASE_URL =
	import.meta.env.VITE_API_URL ?? (import.meta.env.PROD ? "" : "http://localhost:8080");

/** The error body the Go handlers send on every failure. */
type ErrorBody = { error: string; code: string };

/**
 * A failure the server described. The code lets callers branch without
 * matching on message text: "inconsistent_history" in particular means the
 * relayed colours contradict each other rather than that anything broke.
 */
export class ApiError extends Error {
	readonly code: string;
	readonly status: number;

	constructor(message: string, code: string, status: number) {
		super(message);
		this.name = "ApiError";
		this.code = code;
		this.status = status;
	}
}

/** True when a rejection is a deliberate abort rather than a real failure. */
export function isAbort(err: unknown): boolean {
	return err instanceof DOMException && err.name === "AbortError";
}

async function toError(res: Response): Promise<ApiError> {
	try {
		const body: ErrorBody = await res.json();
		return new ApiError(body.error, body.code, res.status);
	} catch {
		// A non-JSON body means something upstream of the handlers failed.
		return new ApiError(`server returned ${res.status}`, "unexpected", res.status);
	}
}

async function request<T>(path: string, init: RequestInit): Promise<T> {
	let res: Response;
	try {
		res = await fetch(`${BASE_URL}${path}`, init);
	} catch (err) {
		if (isAbort(err)) throw err;
		// fetch rejects on network failure and on CORS-blocked responses,
		// which from here are indistinguishable from the server being down.
		throw new ApiError(
			err instanceof Error ? err.message : "request failed",
			"network",
			0,
		);
	}

	if (!res.ok) throw await toError(res);
	return (await res.json()) as T;
}

export function getJSON<T>(path: string, signal?: AbortSignal): Promise<T> {
	return request<T>(path, { signal });
}

export function postJSON<T>(path: string, body: unknown, signal?: AbortSignal): Promise<T> {
	return request<T>(path, {
		method: "POST",
		headers: { "Content-Type": "application/json" },
		body: JSON.stringify(body),
		signal,
	});
}
