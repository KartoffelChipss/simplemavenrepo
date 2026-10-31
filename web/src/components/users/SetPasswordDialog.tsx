import type { User } from '@/api/types.ts';
import { useUpdateUser } from '@/api/users.ts';
import { passwordError } from '@/components/users/validation.ts';
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
import { Field, FieldError, FieldGroup, FieldLabel } from '@/components/ui/field';
import { Input } from '@/components/ui/input';
import { Spinner } from '@/components/ui/spinner';
import { toast } from '@/components/ui/toast';
import { useState, type FormEvent } from 'react';

function SetPasswordForm({ user, onDone }: { user: User; onDone: () => void }) {
    const [password, setPassword] = useState('');
    const [confirm, setConfirm] = useState('');
    const update = useUpdateUser();

    const passwordInvalid = passwordError(password);
    const mismatch = confirm !== '' && confirm !== password;
    const canSubmit = !update.isPending && password && !passwordInvalid && confirm === password;

    function onSubmit(e: FormEvent) {
        e.preventDefault();
        update.mutate(
            { username: user.username, password },
            {
                onSuccess: () => {
                    toast.add({ title: 'Password changed', description: user.username });
                    onDone();
                },
            }
        );
    }

    return (
        <form onSubmit={onSubmit}>
            <DialogHeader>
                <DialogTitle>{user.hasPassword ? 'Change password' : 'Set password'}</DialogTitle>
                <DialogDescription>
                    {user.hasPassword
                        ? `${user.username} is logged out of the web UI everywhere. Their tokens keep working.`
                        : `${user.username} can then log in to the web UI.`}
                </DialogDescription>
            </DialogHeader>
            <FieldGroup className="py-4">
                <Field data-invalid={!!passwordInvalid}>
                    <FieldLabel htmlFor="new-password">New password</FieldLabel>
                    <Input
                        id="new-password"
                        type="password"
                        autoComplete="new-password"
                        value={password}
                        onChange={(e) => setPassword(e.target.value)}
                        aria-invalid={!!passwordInvalid}
                        required
                        autoFocus
                    />
                    {passwordInvalid && <FieldError>{passwordInvalid}</FieldError>}
                </Field>
                <Field data-invalid={mismatch}>
                    <FieldLabel htmlFor="confirm-password">Confirm password</FieldLabel>
                    <Input
                        id="confirm-password"
                        type="password"
                        autoComplete="new-password"
                        value={confirm}
                        onChange={(e) => setConfirm(e.target.value)}
                        aria-invalid={mismatch}
                        required
                    />
                    {mismatch && <FieldError>The passwords don’t match</FieldError>}
                </Field>
                {update.error && <FieldError>{update.error.message}</FieldError>}
            </FieldGroup>
            <DialogFooter>
                <DialogClose render={<Button variant="outline" />}>Cancel</DialogClose>
                <Button type="submit" disabled={!canSubmit}>
                    {update.isPending && <Spinner />}
                    Save password
                </Button>
            </DialogFooter>
        </form>
    );
}

export function SetPasswordDialog({ user }: { user: User }) {
    const [open, setOpen] = useState(false);

    return (
        <Dialog open={open} onOpenChange={setOpen}>
            <DialogTrigger render={<Button variant="outline" size="sm" />}>
                {user.hasPassword ? 'Change password' : 'Set password'}
            </DialogTrigger>
            <DialogContent className="sm:max-w-md">
                <SetPasswordForm user={user} onDone={() => setOpen(false)} />
            </DialogContent>
        </Dialog>
    );
}
