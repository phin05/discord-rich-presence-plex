export const apiBaseUrl = import.meta.env.DEV ? "http://localhost:8040" : "";

const fetchTimeout = 10000;

export interface ApiError extends Error {
	details?: string[];
}

export async function apiRequest<T>(method: string, url: string, body?: unknown) {
	if (!url.startsWith("https://") && !url.startsWith("http://")) {
		url = `${apiBaseUrl}/api/${url}`;
	}
	const response = await fetch(url, {
		method,
		signal: AbortSignal.timeout(fetchTimeout),
		body: body !== undefined ? JSON.stringify(body) : undefined,
		headers: body !== undefined ? { "Content-Type": "application/json" } : undefined,
	});
	const isJson = response.headers.get("Content-Type")?.includes("application/json");
	const data = isJson ? ((await response.json()) as unknown) : {};
	if (data && typeof data === "object" && "error" in data) {
		throw data.error;
	}
	if (!response.ok || (!isJson && response.status !== 204)) {
		throw new Error(response.ok ? "Invalid response format" : `HTTP status ${response.status} ${response.statusText}`);
	}
	return data as T;
}
