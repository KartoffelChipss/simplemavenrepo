import { useChangeMyPassword } from '@/api/auth.ts';
import { isApiError } from '@/api/client.ts';
import type { User } from '@/api/types.ts';
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
} from '@/components/ui/dialog';
import { Field, FieldError, FieldGroup, FieldLabel } from '@/components/ui/field';
import { Input } from '@/components/ui/input';
import { Spinner } from '@/components/ui/spinner';
import { toast } from '@/components/ui/toast';
import { useState, type FormEvent } from 'react';

function ChangeMyPasswordForm({ user, onDone }: { user: User; onDone: () => void }) {
    const [current, setCurrent] = useState('');
    const [password, setPassword] = useState('');
    const [confirm, setConfirm] = useState('');
    const change = useChangeMyPassword();

    const passwordInvalid = passwordError(password);
    const mismatch = confirm !== '' && confirm !== password;
    // the server answers 403 if the current password is wrong
    const currentWrong = isApiError(change.error, 403);
    const canSubmit =
        !change.isPending && (!user.hasPassword || current) && password && !passwordInvalid && confirm === password;

    function onSubmit(e: FormEvent) {
        e.preventDefault();
        change.mutate(
            { currentPassword: user.hasPassword ? current : undefined, newPassword: password },
            {
                onSuccess: () => {
                    toast.add({ title: 'Password changed', description: 'You were logged out everywhere else.' });
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
                    Other web UI sessions are logged out, this one stays logged in. Your tokens keep working.
                </DialogDescription>
            </DialogHeader>
            <FieldGroup className="py-4">
                {/* lets password managers attach the new password to the right account */}
                <input type="text" autoComplete="username" value={user.username} hidden readOnly />
                {user.hasPassword && (
                    <Field data-invalid={currentWrong}>
                        <FieldLabel htmlFor="current-password">Current password</FieldLabel>
                        <Input
                            id="current-password"
                            type="password"
                            autoComplete="current-password"
                            value={current}
                            onChange={(e) => {
                                setCurrent(e.target.value);
                                if (currentWrong) change.reset();
                            }}
                            aria-invalid={currentWrong}
                            required
                            autoFocus
                        />
                        {currentWrong && <FieldError>The current password is wrong</FieldError>}
                    </Field>
                )}
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
                        autoFocus={!user.hasPassword}
                    />
                    {passwordInvalid && <FieldError>{passwordInvalid}</FieldError>}
                </Field>
                <Field data-invalid={mismatch}>
                    <FieldLabel htmlFor="confirm-password">Confirm new password</FieldLabel>
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
                {change.error && !currentWrong && <FieldError>{change.error.message}</FieldError>}
            </FieldGroup>
            <DialogFooter>
                <DialogClose render={<Button variant="outline" />}>Cancel</DialogClose>
                <Button type="submit" disabled={!canSubmit}>
                    {change.isPending && <Spinner />}
                    Save password
                </Button>
            </DialogFooter>
        </form>
    );
}

export function ChangeMyPasswordDialog({
    user,
    open,
    onOpenChange,
}: {
    user: User;
    open: boolean;
    onOpenChange: (open: boolean) => void;
}) {
    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <DialogContent className="sm:max-w-md">
                <ChangeMyPasswordForm user={user} onDone={() => onOpenChange(false)} />
            </DialogContent>
        </Dialog>
    );
}
