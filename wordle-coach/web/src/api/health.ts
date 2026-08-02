const BASE_URL = import.meta.env.VITE_API_URL ?? "http://localhost:8080";

/** Shape of the /healthz response body. */
type HealthResponse = { status: string };

/** Either the reported status, or why the check could not be completed. */
export type HealthResult =
	| { ok: true; status: string }
	| { ok: false; error: string };

export async function checkHealth(): Promise<HealthResult> {
	try {
		const res = await fetch(`${BASE_URL}/healthz`);

		if (!res.ok) {
			return { ok: false, error: `server returned ${res.status}` };
		}

		const data: HealthResponse = await res.json();
		return { ok: true, status: data.status };
	} catch (err) {
		// fetch rejects on network failure and on CORS-blocked responses.
		return {
			ok: false,
			error: err instanceof Error ? err.message : "request failed",
		};
	}
}
