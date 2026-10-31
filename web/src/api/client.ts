export class ApiError extends Error {
    readonly status: number;

    constructor(status: number, message: string) {
        super(message);
        this.name = 'ApiError';
        this.status = status;
    }
}

export function isApiError(err: unknown, status?: number): err is ApiError {
    return err instanceof ApiError && (status === undefined || err.status === status);
}

type Method = 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE';

/**
 * Calls the management API under /api. Bodies are sent and received as JSON. The browser sends the HttpOnly session
 * cookie set by the login on its own.
 */
export async function request<T>(method: Method, path: string, body?: unknown, signal?: AbortSignal): Promise<T> {
    const res = await fetch(`/api${path}`, {
        method,
        signal,
        headers: body === undefined ? undefined : { 'Content-Type': 'application/json' },
        body: body === undefined ? undefined : JSON.stringify(body),
    });
    if (!res.ok) {
        const error = (await res.json().catch(() => null)) as { error?: string } | null;
        throw new ApiError(res.status, error?.error ?? res.statusText);
    }
    if (res.status === 204) return undefined as T;
    return res.json() as Promise<T>;
}

/** Encodes a value for use as one path segment. */
export const seg = encodeURIComponent;
