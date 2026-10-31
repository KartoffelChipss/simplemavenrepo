import { useCallback, useEffect, useState, type ReactNode } from 'react';
import { api, getToken, setToken, type User } from './api.ts';
import { AuthContext } from './auth-context.ts';

export function AuthProvider({ children }: { children: ReactNode }) {
    const [user, setUser] = useState<User | null>(null);
    const [loading, setLoading] = useState(() => getToken() !== null);

    useEffect(() => {
        if (!getToken()) return;
        api<User>('/me')
            .then(setUser)
            .catch(() => setToken(null))
            .finally(() => setLoading(false));
    }, []);

    const login = useCallback(async (token: string) => {
        const me = await api<User>('/me', {}, token);
        setToken(token);
        setUser(me);
    }, []);

    const logout = useCallback(() => {
        setToken(null);
        setUser(null);
    }, []);

    return <AuthContext value={{ user, loading, login, logout }}>{children}</AuthContext>;
}
