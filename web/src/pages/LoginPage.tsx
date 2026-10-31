import { useState, type FormEvent } from 'react';
import { Navigate, useLocation } from 'react-router';
import { useLogin, useMe } from '@/api/auth.ts';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Field, FieldError, FieldGroup, FieldLabel } from '@/components/ui/field';
import { Input } from '@/components/ui/input';
import { Spinner } from '@/components/ui/spinner';
import { useDocumentTitle } from '@/title.ts';
import { Package } from 'lucide-react';

export function LoginPage() {
    const { data: user } = useMe();
    const login = useLogin();
    const location = useLocation();
    const from = (location.state as { from?: string } | null)?.from ?? '/';

    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');
    useDocumentTitle('Log in');

    if (user) return <Navigate to={from} replace />;

    function onSubmit(e: FormEvent) {
        e.preventDefault();
        // Once logged in, the user is set and the redirect above takes over.
        login.mutate({ username: username.trim(), password });
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
                        <form onSubmit={onSubmit}>
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
                                {login.error && <FieldError>{login.error.message}</FieldError>}
                                <Button type="submit" disabled={login.isPending || !username.trim() || !password}>
                                    {login.isPending && <Spinner />}
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
