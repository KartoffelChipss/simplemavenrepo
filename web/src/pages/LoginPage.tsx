import { useState, type FormEvent } from 'react';
import { Navigate, useLocation, useNavigate } from 'react-router';
import { useAuth } from '@/auth-context.ts';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Field, FieldError, FieldGroup, FieldLabel } from '@/components/ui/field';
import { Input } from '@/components/ui/input';
import { Spinner } from '@/components/ui/spinner';
import { Package } from 'lucide-react';

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
        <div className="flex min-h-svh items-center justify-center bg-muted p-6">
            <div className="flex w-full max-w-sm flex-col gap-6">
                <div className="flex items-center gap-2 self-center font-medium">
                    <div className="flex size-7 items-center justify-center rounded-md bg-primary text-primary-foreground">
                        <Package className="size-4" />
                    </div>
                    simplemavenrepo
                </div>
                <Card>
                    <CardHeader>
                        <CardTitle>Log in</CardTitle>
                        <CardDescription>Use your username and password to manage the repository.</CardDescription>
                    </CardHeader>
                    <CardContent>
                        <form onSubmit={(e) => void onSubmit(e)}>
                            <FieldGroup>
                                <Field>
                                    <FieldLabel htmlFor="username">Username</FieldLabel>
                                    <Input
                                        id="username"
                                        autoComplete="username"
                                        autoCapitalize="none"
                                        value={username}
                                        onChange={(e) => setUsername(e.target.value)}
                                        required
                                        autoFocus
                                    />
                                </Field>
                                <Field>
                                    <FieldLabel htmlFor="password">Password</FieldLabel>
                                    <Input
                                        id="password"
                                        type="password"
                                        autoComplete="current-password"
                                        value={password}
                                        onChange={(e) => setPassword(e.target.value)}
                                        required
                                    />
                                </Field>
                                {error && <FieldError>{error}</FieldError>}
                                <Button type="submit" disabled={submitting || !username.trim() || !password}>
                                    {submitting && <Spinner />}
                                    Log in
                                </Button>
                            </FieldGroup>
                        </form>
                    </CardContent>
                </Card>
            </div>
        </div>
    );
}
