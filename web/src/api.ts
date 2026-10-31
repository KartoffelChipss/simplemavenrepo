const TOKEN_KEY = 'smr.token';

export function getToken(): string | null {
    try {
        return localStorage.getItem(TOKEN_KEY);
    } catch {
        return null;
    }
}

export function setToken(token: string | null) {
    try {
        if (token) localStorage.setItem(TOKEN_KEY, token);
        else localStorage.removeItem(TOKEN_KEY);
    } catch {
        // Storage unavailable, the token only lives for this page load.
    }
}

export class ApiError extends Error {
    readonly status: number;

    constructor(status: number, message: string) {
        super(message);
        this.status = status;
    }
}

/** Calls the management API, sending the stored token as a bearer token. */
export async function api<T>(path: string, init: RequestInit = {}, token = getToken()): Promise<T> {
    const headers = new Headers(init.headers);
    if (token) headers.set('Authorization', `Bearer ${token}`);
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
    createdAt: string;
}
