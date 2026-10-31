import { useState, type FormEvent } from 'react';
import { Navigate, useLocation, useNavigate } from 'react-router';
import { useAuth } from '../auth-context.ts';

export function LoginPage() {
    const { user, login } = useAuth();
    const navigate = useNavigate();
    const location = useLocation();
    const from = (location.state as { from?: string } | null)?.from ?? '/';

    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');
    const [error, setError] = useState<string | null>(null);
    const [submitting, setSubmitting] = useState(false);

    if (user) return <Navigate to={from} replace />;

    async function onSubmit(e: FormEvent) {
        e.preventDefault();
        setSubmitting(true);
        setError(null);
        try {
            await login(username.trim(), password);
            await navigate(from, { replace: true });
        } catch (err) {
            setError(err instanceof Error ? err.message : String(err));
        } finally {
            setSubmitting(false);
        }
    }

    return (
        <div className="login">
            <h1>simplemavenrepo</h1>
            <form onSubmit={(e) => void onSubmit(e)}>
                <label htmlFor="username">Username</label>
                <input
                    id="username"
                    autoComplete="username"
                    autoCapitalize="none"
                    value={username}
                    onChange={(e) => setUsername(e.target.value)}
                    required
                    autoFocus
                />
                <label htmlFor="password">Password</label>
                <input
                    id="password"
                    type="password"
                    autoComplete="current-password"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    required
                />
                {error && <p className="error">{error}</p>}
                <button type="submit" disabled={submitting || !username.trim() || !password}>
                    Log in
                </button>
            </form>
        </div>
    );
}
