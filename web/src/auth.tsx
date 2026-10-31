import { useCallback, useEffect, useState, type ReactNode } from 'react';
import { api, type User } from './api.ts';
import { AuthContext } from './auth-context.ts';

export function AuthProvider({ children }: { children: ReactNode }) {
    const [user, setUser] = useState<User | null>(null);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        api<User>('/me')
            .then(setUser)
            .catch(() => setUser(null))
            .finally(() => setLoading(false));
    }, []);

    const login = useCallback(async (username: string, password: string) => {
        const me = await api<User>('/auth/login', {
            method: 'POST',
            body: JSON.stringify({ username, password }),
        });
        setUser(me);
    }, []);

    const logout = useCallback(async () => {
        await api('/auth/logout', { method: 'POST' });
        setUser(null);
    }, []);

    return <AuthContext value={{ user, loading, login, logout }}>{children}</AuthContext>;
}
