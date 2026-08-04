import { ApiError, getJSON } from "./client";

/** Shape of the /healthz response body. */
type HealthResponse = { status: string };

/** Either the reported status, or why the check could not be completed. */
export type HealthResult =
	| { ok: true; status: string }
	| { ok: false; error: string };

export async function checkHealth(): Promise<HealthResult> {
	try {
		const data = await getJSON<HealthResponse>("/healthz");
		return { ok: true, status: data.status };
	} catch (err) {
		// ApiError already covers network failures and CORS-blocked responses,
		// which from the browser are indistinguishable from the server being down.
		return {
			ok: false,
			error: err instanceof ApiError || err instanceof Error ? err.message : "request failed",
		};
	}
}
