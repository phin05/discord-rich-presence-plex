export const apiBaseUrl = import.meta.env.DEV ? "http://localhost:8040" : "";

const fetchTimeout = 10000;

export interface ApiError {
	message: string;
	details?: string[];
}

interface ApiErrorResponse {
	error?: ApiError;
}

export async function apiRequest<T>(method: string, url: string, body?: unknown) {
	let data: ApiErrorResponse;
	try {
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
		data = isJson ? ((await response.json()) as ApiErrorResponse) : {};
		if (!data.error && (!response.ok || (!isJson && response.status !== 204))) {
			data = { error: { message: response.ok ? "Invalid response format" : `HTTP status ${response.status} ${response.statusText}` } };
		}
	} catch (error) {
		data = { error: { message: (error as Error).message } };
	}
	if (data.error) {
		// eslint-disable-next-line @typescript-eslint/only-throw-error
		throw data.error;
	}
	return data as T;
}
