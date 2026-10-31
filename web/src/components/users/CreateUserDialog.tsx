import { useCreateUser } from '@/api/users.ts';
import { passwordError, usernameError } from '@/components/users/validation.ts';
import { Button } from '@/components/ui/button';
import {
    Dialog,
    DialogClose,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
} from '@/components/ui/dialog';
import { Field, FieldContent, FieldDescription, FieldError, FieldGroup, FieldLabel } from '@/components/ui/field';
import { Input } from '@/components/ui/input';
import { Spinner } from '@/components/ui/spinner';
import { Switch } from '@/components/ui/switch';
import { toast } from '@/components/ui/toast';
import { Plus } from 'lucide-react';
import { useState, type FormEvent } from 'react';
import { useNavigate } from 'react-router';

function CreateUserForm() {
    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');
    const [isAdmin, setIsAdmin] = useState(false);
    const create = useCreateUser();
    const navigate = useNavigate();

    const usernameInvalid = usernameError(username.trim());
    const passwordInvalid = passwordError(password);
    const canSubmit = !create.isPending && username.trim() && !usernameInvalid && !passwordInvalid;

    function onSubmit(e: FormEvent) {
        e.preventDefault();
        create.mutate(
            { username: username.trim(), isAdmin, password: password || undefined },
            {
                onSuccess: (user) => {
                    toast.add({ title: 'User created', description: user.username });
                    void navigate(`/users/${encodeURIComponent(user.username)}`);
                },
            }
        );
    }

    return (
        <form onSubmit={onSubmit}>
            <DialogHeader>
                <DialogTitle>Create user</DialogTitle>
                <DialogDescription>
                    Users authenticate with tokens in their build tools. A password lets them log in here as well.
                </DialogDescription>
            </DialogHeader>
            <FieldGroup className="py-4">
                <Field data-invalid={!!usernameInvalid}>
                    <FieldLabel htmlFor="user-name">Username</FieldLabel>
                    <Input
                        id="user-name"
                        autoComplete="off"
                        autoCapitalize="none"
                        spellCheck={false}
                        value={username}
                        onChange={(e) => setUsername(e.target.value)}
                        aria-invalid={!!usernameInvalid}
                        required
                        autoFocus
                    />
                    {usernameInvalid && <FieldError>{usernameInvalid}</FieldError>}
                </Field>
                <Field data-invalid={!!passwordInvalid}>
                    <FieldLabel htmlFor="user-password">Password</FieldLabel>
                    <Input
                        id="user-password"
                        type="password"
                        autoComplete="new-password"
                        placeholder="Optional"
                        value={password}
                        onChange={(e) => setPassword(e.target.value)}
                        aria-invalid={!!passwordInvalid}
                    />
                    {passwordInvalid ? (
                        <FieldError>{passwordInvalid}</FieldError>
                    ) : (
                        <FieldDescription>Without one, the user can only use tokens.</FieldDescription>
                    )}
                </Field>
                <Field orientation="horizontal">
                    <FieldContent>
                        <FieldLabel id="user-admin-label" htmlFor="user-admin">
                            Administrator
                        </FieldLabel>
                        <FieldDescription>
                            Can manage users, teams, grants and artifacts, and access everything.
                        </FieldDescription>
                    </FieldContent>
                    <Switch id="user-admin" checked={isAdmin} onCheckedChange={setIsAdmin} />
                </Field>
                {create.error && <FieldError>{create.error.message}</FieldError>}
            </FieldGroup>
            <DialogFooter>
                <DialogClose render={<Button variant="outline" />}>Cancel</DialogClose>
                <Button type="submit" disabled={!canSubmit}>
                    {create.isPending && <Spinner />}
                    Create user
                </Button>
            </DialogFooter>
        </form>
    );
}

export function CreateUserDialog() {
    return (
        <Dialog>
            <DialogTrigger render={<Button />}>
                <Plus />
                Create user
            </DialogTrigger>
            <DialogContent className="sm:max-w-md">
                {/* Mounted only while open, so the form starts fresh every time. */}
                <CreateUserForm />
            </DialogContent>
        </Dialog>
    );
}
