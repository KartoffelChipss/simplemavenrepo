import { useState, type FormEvent } from 'react';
import { Navigate, useLocation, useNavigate } from 'react-router';
import { useAuth } from '../auth-context.ts';

export function LoginPage() {
    const { user, login } = useAuth();
    const navigate = useNavigate();
    const location = useLocation();
    const from = (location.state as { from?: string } | null)?.from ?? '/';

    const [token, setToken] = useState('');
    const [error, setError] = useState<string | null>(null);
    const [submitting, setSubmitting] = useState(false);

    if (user) return <Navigate to={from} replace />;

    async function onSubmit(e: FormEvent) {
        e.preventDefault();
        setSubmitting(true);
        setError(null);
        try {
            await login(token.trim());
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
                <label htmlFor="token">Access token</label>
                <input
                    id="token"
                    type="password"
                    autoComplete="current-password"
                    value={token}
                    onChange={(e) => setToken(e.target.value)}
                    required
                />
                {error && <p className="error">{error}</p>}
                <button type="submit" disabled={submitting || !token.trim()}>
                    Log in
                </button>
            </form>
        </div>
    );
}
