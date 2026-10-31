export class ApiError extends Error {
    readonly status: number;

    constructor(status: number, message: string) {
        super(message);
        this.status = status;
    }
}

/** Calls the management API. The browser sends the HttpOnly session cookie set by the login. */
export async function api<T>(path: string, init: RequestInit = {}): Promise<T> {
    const headers = new Headers(init.headers);
    if (init.body && !headers.has('Content-Type')) headers.set('Content-Type', 'application/json');

    const res = await fetch(`/api${path}`, { ...init, headers });
    if (!res.ok) {
        const body = (await res.json().catch(() => null)) as { error?: string } | null;
        throw new ApiError(res.status, body?.error ?? res.statusText);
    }
    if (res.status === 204) return undefined as T;
    return res.json() as Promise<T>;
}

export interface User {
    id: number;
    username: string;
    isAdmin: boolean;
    hasPassword: boolean;
    createdAt: string;
}
