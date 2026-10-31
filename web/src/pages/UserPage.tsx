import { useMe } from '@/api/auth.ts';
import { isApiError } from '@/api/client.ts';
import { useUserTokens } from '@/api/tokens.ts';
import type { User, UserDetails } from '@/api/types.ts';
import { useDeleteUser, useUpdateUser, useUser } from '@/api/users.ts';
import { GrantEditor, PermissionBadge } from '@/components/grants/GrantEditor.tsx';
import { PageHeader } from '@/components/PageHeader.tsx';
import { CreateTokenDialog } from '@/components/tokens/CreateTokenDialog.tsx';
import { TokenTable } from '@/components/tokens/TokenTable.tsx';
import { SetPasswordDialog } from '@/components/users/SetPasswordDialog.tsx';
import {
    AlertDialog,
    AlertDialogAction,
    AlertDialogCancel,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
    AlertDialogTrigger,
} from '@/components/ui/alert-dialog';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Field, FieldContent, FieldDescription, FieldLabel } from '@/components/ui/field';
import { Separator } from '@/components/ui/separator';
import { Skeleton } from '@/components/ui/skeleton';
import { Switch } from '@/components/ui/switch';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { toast } from '@/components/ui/toast';
import { formatDate } from '@/lib/format.ts';
import { ArrowLeft } from 'lucide-react';
import { usePageTitle } from '@/title.ts';
import { Link, useNavigate, useParams } from 'react-router';

function AdminSwitch({ user, isSelf }: { user: User; isSelf: boolean }) {
    const update = useUpdateUser();

    function onChange(isAdmin: boolean) {
        update.mutate(
            { username: user.username, isAdmin },
            {
                onSuccess: () =>
                    toast.add({
                        title: isAdmin ? 'Made administrator' : 'Removed administrator',
                        description: user.username,
                    }),
                // e.g. the server refuses to demote the last admin
                onError: (err) => toast.add({ title: 'Could not change role', description: err.message }),
            }
        );
    }

    return (
        <Field orientation="horizontal">
            <FieldContent>
                <FieldLabel id="user-admin-label" htmlFor="user-admin">
                    Administrator
                </FieldLabel>
                <FieldDescription>
                    {isSelf
                        ? 'You can’t remove your own admin access, ask another administrator.'
                        : 'Can manage users, teams, grants and artifacts, and access everything.'}
                </FieldDescription>
            </FieldContent>
            <Switch
                id="user-admin"
                checked={user.isAdmin}
                onCheckedChange={onChange}
                disabled={isSelf || update.isPending}
            />
        </Field>
    );
}

function RemovePasswordButton({ user }: { user: User }) {
    const update = useUpdateUser();

    function onRemove() {
        update.mutate(
            { username: user.username, password: '' },
            {
                onSuccess: () => toast.add({ title: 'Password removed', description: user.username }),
                onError: (err) => toast.add({ title: 'Could not remove password', description: err.message }),
            }
        );
    }

    return (
        <AlertDialog>
            <AlertDialogTrigger render={<Button variant="ghost" size="sm" />}>Remove</AlertDialogTrigger>
            <AlertDialogContent>
                <AlertDialogHeader>
                    <AlertDialogTitle>Remove {user.username}’s password?</AlertDialogTitle>
                    <AlertDialogDescription>
                        They are logged out of the web UI and can’t log in again until a password is set. Their tokens
                        keep working.
                    </AlertDialogDescription>
                </AlertDialogHeader>
                <AlertDialogFooter>
                    <AlertDialogCancel>Cancel</AlertDialogCancel>
                    <AlertDialogAction variant="destructive" disabled={update.isPending} onClick={onRemove}>
                        Remove password
                    </AlertDialogAction>
                </AlertDialogFooter>
            </AlertDialogContent>
        </AlertDialog>
    );
}

function AccountCard({ user, isSelf }: { user: User; isSelf: boolean }) {
    return (
        <Card>
            <CardHeader>
                <CardTitle>Account</CardTitle>
                <CardDescription>Created {formatDate(user.createdAt)}</CardDescription>
            </CardHeader>
            <CardContent className="flex flex-col gap-4">
                <AdminSwitch user={user} isSelf={isSelf} />
                <Separator />
                <div className="flex flex-wrap items-center justify-between gap-2">
                    <div className="flex flex-col gap-1">
                        <span className="text-sm font-medium">Web UI password</span>
                        <span className="text-sm text-muted-foreground">
                            {user.hasPassword
                                ? 'Set. The user can log in here.'
                                : 'Not set. The user can only use tokens.'}
                        </span>
                    </div>
                    <div className="flex gap-2">
                        {user.hasPassword && !isSelf && <RemovePasswordButton user={user} />}
                        <SetPasswordDialog user={user} />
                    </div>
                </div>
            </CardContent>
        </Card>
    );
}

const teamPath = (name: string) => `/teams/${encodeURIComponent(name)}`;

function AccessCard({ user }: { user: UserDetails }) {
    const direct = user.grants.filter((g) => g.user !== undefined);
    const fromTeams = user.grants.filter((g) => g.team !== undefined);

    return (
        <Card>
            <CardHeader>
                <CardTitle>Access</CardTitle>
                <CardDescription>
                    {user.isAdmin
                        ? 'Administrators can read and write everything, grants don’t limit them.'
                        : 'What this user can read and publish, directly or through their teams.'}
                </CardDescription>
            </CardHeader>
            <CardContent className="flex flex-col gap-6">
                <div className="flex flex-wrap items-center gap-2">
                    <span className="text-sm font-medium">Teams</span>
                    {user.teams.length === 0 && <span className="text-sm text-muted-foreground">None</span>}
                    {user.teams.map((team) => (
                        <Badge key={team.id} variant="secondary" render={<Link to={teamPath(team.name)} />}>
                            {team.name}
                        </Badge>
                    ))}
                </div>
                <div className="flex flex-col gap-3">
                    <h3 className="text-sm font-medium">Direct grants</h3>
                    <GrantEditor grants={direct} target={{ user: user.username }} />
                </div>
                {fromTeams.length > 0 && (
                    <div className="flex flex-col gap-3">
                        <h3 className="text-sm font-medium">From teams</h3>
                        <div className="rounded-lg border">
                            <Table>
                                <TableHeader>
                                    <TableRow>
                                        <TableHead>Repository path</TableHead>
                                        <TableHead>Permission</TableHead>
                                        <TableHead>Team</TableHead>
                                    </TableRow>
                                </TableHeader>
                                <TableBody>
                                    {fromTeams.map((grant) => (
                                        <TableRow key={grant.id}>
                                            <TableCell className="font-mono">{grant.pattern}</TableCell>
                                            <TableCell>
                                                <PermissionBadge permission={grant.permission} />
                                            </TableCell>
                                            <TableCell>
                                                <Link to={teamPath(grant.team!)} className="font-medium">
                                                    {grant.team}
                                                </Link>
                                            </TableCell>
                                        </TableRow>
                                    ))}
                                </TableBody>
                            </Table>
                        </div>
                    </div>
                )}
            </CardContent>
        </Card>
    );
}

function TokensSection({ username }: { username: string }) {
    const { data: tokens, error, isPending } = useUserTokens(username);

    return (
        <section className="flex flex-col gap-4">
            <div className="flex flex-wrap items-end justify-between gap-4">
                <div className="flex flex-col gap-1">
                    <h2 className="text-lg font-semibold">Tokens</h2>
                    <p className="text-sm text-muted-foreground">Used by this user’s build tools and scripts.</p>
                </div>
                <CreateTokenDialog username={username} />
            </div>
            <TokenTable
                tokens={tokens}
                isPending={isPending}
                error={error}
                username={username}
                emptyDescription="This user has no tokens yet."
            />
        </section>
    );
}

function DeleteUserCard({ user, isSelf }: { user: User; isSelf: boolean }) {
    const remove = useDeleteUser();
    const navigate = useNavigate();

    function onDelete() {
        remove.mutate(user.username, {
            onSuccess: () => {
                toast.add({ title: 'User deleted', description: user.username });
                void navigate('/users');
            },
            onError: (err) => toast.add({ title: 'Could not delete user', description: err.message }),
        });
    }

    return (
        <Card className="border-destructive/40">
            <CardHeader>
                <CardTitle>Delete user</CardTitle>
                <CardDescription>
                    {isSelf
                        ? 'You can’t delete your own account, ask another administrator.'
                        : 'Removes the user with their tokens, team memberships and grants. This can’t be undone.'}
                </CardDescription>
            </CardHeader>
            <CardContent>
                <AlertDialog>
                    <AlertDialogTrigger render={<Button variant="destructive" disabled={isSelf} />}>
                        Delete {user.username}
                    </AlertDialogTrigger>
                    <AlertDialogContent>
                        <AlertDialogHeader>
                            <AlertDialogTitle>Delete {user.username}?</AlertDialogTitle>
                            <AlertDialogDescription>
                                Their tokens stop working immediately, and their team memberships and grants are
                                removed. This can’t be undone.
                            </AlertDialogDescription>
                        </AlertDialogHeader>
                        <AlertDialogFooter>
                            <AlertDialogCancel>Cancel</AlertDialogCancel>
                            <AlertDialogAction variant="destructive" disabled={remove.isPending} onClick={onDelete}>
                                Delete user
                            </AlertDialogAction>
                        </AlertDialogFooter>
                    </AlertDialogContent>
                </AlertDialog>
            </CardContent>
        </Card>
    );
}

export function UserPage() {
    const { username = '' } = useParams();
    const { data: user, error, isPending } = useUser(username);
    usePageTitle(username);
    const { data: me } = useMe();
    const isSelf = user !== undefined && user.id === me?.id;

    return (
        <>
            <Link
                to="/users"
                className="flex w-fit items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
            >
                <ArrowLeft className="size-4" />
                Users
            </Link>

            {isPending && (
                <div className="flex flex-col gap-4">
                    <Skeleton className="h-8 w-48" />
                    <Skeleton className="h-40" />
                </div>
            )}

            {isApiError(error, 404) && (
                <PageHeader title="User not found" description={`There is no user “${username}”.`} />
            )}
            {error && !isApiError(error, 404) && <p className="text-destructive">{error.message}</p>}

            {user && (
                <>
                    <div className="flex flex-wrap items-center gap-3">
                        <h1 className="text-2xl font-semibold tracking-tight">{user.username}</h1>
                        {user.isAdmin && <Badge>Administrator</Badge>}
                        {isSelf && <Badge variant="outline">You</Badge>}
                    </div>
                    <div className="grid gap-4 xl:grid-cols-2">
                        <AccountCard user={user} isSelf={isSelf} />
                        <AccessCard user={user} />
                    </div>
                    <TokensSection username={user.username} />
                    <DeleteUserCard user={user} isSelf={isSelf} />
                </>
            )}
        </>
    );
}
