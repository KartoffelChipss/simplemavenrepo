import { isApiError } from '@/api/client.ts';
import { useAddTeamMember, useDeleteTeam, useRemoveTeamMember, useTeam } from '@/api/teams.ts';
import type { TeamDetails, User } from '@/api/types.ts';
import { useUsers } from '@/api/users.ts';
import { GrantEditor } from '@/components/grants/GrantEditor.tsx';
import { PageHeader } from '@/components/PageHeader.tsx';
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
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Skeleton } from '@/components/ui/skeleton';
import { Spinner } from '@/components/ui/spinner';
import { Table, TableBody, TableCell, TableRow } from '@/components/ui/table';
import { toast } from '@/components/ui/toast';
import { formatDate } from '@/lib/format.ts';
import { ArrowLeft, Plus, UserMinus } from 'lucide-react';
import { useState, type FormEvent } from 'react';
import { usePageTitle } from '@/title.ts';
import { Link, useNavigate, useParams } from 'react-router';

function RemoveMemberButton({ team, user }: { team: string; user: User }) {
    const remove = useRemoveTeamMember();

    function onRemove() {
        remove.mutate(
            { team, username: user.username },
            {
                onSuccess: () => toast.add({ title: 'Member removed', description: `${user.username} left ${team}` }),
                onError: (err) => toast.add({ title: 'Could not remove member', description: err.message }),
            }
        );
    }

    return (
        <Button
            variant="ghost"
            size="icon-sm"
            aria-label={`Remove ${user.username} from ${team}`}
            disabled={remove.isPending}
            onClick={onRemove}
        >
            <UserMinus />
        </Button>
    );
}

function AddMemberForm({ team }: { team: TeamDetails }) {
    const { data: users } = useUsers();
    const [username, setUsername] = useState<string | null>(null);
    const add = useAddTeamMember();

    const memberIds = new Set(team.members.map((m) => m.id));
    const candidates = (users ?? [])
        .filter((u) => !memberIds.has(u.id))
        .map((u) => ({ value: u.username, label: u.username }));

    function onSubmit(e: FormEvent) {
        e.preventDefault();
        if (!username) return;
        add.mutate(
            { team: team.name, username },
            {
                onSuccess: () => {
                    toast.add({ title: 'Member added', description: `${username} joined ${team.name}` });
                    setUsername(null);
                },
                onError: (err) => toast.add({ title: 'Could not add member', description: err.message }),
            }
        );
    }

    if (users && candidates.length === 0) {
        return <p className="text-sm text-muted-foreground">Every user is already a member.</p>;
    }

    return (
        <form onSubmit={onSubmit} className="flex flex-wrap gap-2">
            <Select items={candidates} value={username} onValueChange={setUsername}>
                <SelectTrigger aria-label="User to add" className="min-w-48 flex-1">
                    <SelectValue placeholder="Choose a user…" />
                </SelectTrigger>
                <SelectContent>
                    {candidates.map((c) => (
                        <SelectItem key={c.value} value={c.value}>
                            {c.label}
                        </SelectItem>
                    ))}
                </SelectContent>
            </Select>
            <Button type="submit" disabled={!username || add.isPending}>
                {add.isPending ? <Spinner /> : <Plus />}
                Add member
            </Button>
        </form>
    );
}

function MembersCard({ team }: { team: TeamDetails }) {
    return (
        <Card>
            <CardHeader>
                <CardTitle>Members</CardTitle>
                <CardDescription>Members get all of the team’s grants.</CardDescription>
            </CardHeader>
            <CardContent className="flex flex-col gap-4">
                {team.members.length === 0 ? (
                    <p className="text-sm text-muted-foreground">No members yet.</p>
                ) : (
                    <div className="rounded-lg border">
                        <Table>
                            <TableBody>
                                {team.members.map((member) => (
                                    <TableRow key={member.id}>
                                        <TableCell className="font-medium">
                                            <Link to={`/users/${encodeURIComponent(member.username)}`}>
                                                {member.username}
                                            </Link>
                                            {member.isAdmin && (
                                                <Badge variant="secondary" className="ml-2">
                                                    Administrator
                                                </Badge>
                                            )}
                                        </TableCell>
                                        <TableCell className="w-0">
                                            <RemoveMemberButton team={team.name} user={member} />
                                        </TableCell>
                                    </TableRow>
                                ))}
                            </TableBody>
                        </Table>
                    </div>
                )}
                <AddMemberForm team={team} />
            </CardContent>
        </Card>
    );
}

function DeleteTeamCard({ team }: { team: TeamDetails }) {
    const remove = useDeleteTeam();
    const navigate = useNavigate();

    function onDelete() {
        remove.mutate(team.name, {
            onSuccess: () => {
                toast.add({ title: 'Team deleted', description: team.name });
                void navigate('/teams');
            },
            onError: (err) => toast.add({ title: 'Could not delete team', description: err.message }),
        });
    }

    return (
        <Card className="border-destructive/40">
            <CardHeader>
                <CardTitle>Delete team</CardTitle>
                <CardDescription>
                    Removes the team and its grants. The members keep their accounts and their own grants.
                </CardDescription>
            </CardHeader>
            <CardContent>
                <AlertDialog>
                    <AlertDialogTrigger render={<Button variant="destructive" />}>
                        Delete {team.name}
                    </AlertDialogTrigger>
                    <AlertDialogContent>
                        <AlertDialogHeader>
                            <AlertDialogTitle>Delete {team.name}?</AlertDialogTitle>
                            <AlertDialogDescription>
                                Its {team.members.length === 1 ? 'member loses' : 'members lose'} the access the team
                                gave them. This can’t be undone.
                            </AlertDialogDescription>
                        </AlertDialogHeader>
                        <AlertDialogFooter>
                            <AlertDialogCancel>Cancel</AlertDialogCancel>
                            <AlertDialogAction variant="destructive" disabled={remove.isPending} onClick={onDelete}>
                                Delete team
                            </AlertDialogAction>
                        </AlertDialogFooter>
                    </AlertDialogContent>
                </AlertDialog>
            </CardContent>
        </Card>
    );
}

export function TeamPage() {
    const { name = '' } = useParams();
    const { data: team, error, isPending } = useTeam(name);
    usePageTitle(name);

    return (
        <>
            <Link
                to="/teams"
                className="flex w-fit items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
            >
                <ArrowLeft className="size-4" />
                Teams
            </Link>

            {isPending && (
                <div className="flex flex-col gap-4">
                    <Skeleton className="h-8 w-48" />
                    <Skeleton className="h-40" />
                </div>
            )}

            {isApiError(error, 404) && (
                <PageHeader title="Team not found" description={`There is no team “${name}”.`} />
            )}
            {error && !isApiError(error, 404) && <p className="text-destructive">{error.message}</p>}

            {team && (
                <>
                    <PageHeader
                        title={team.name}
                        description={`Created ${formatDate(team.createdAt)} · ${team.members.length} ${
                            team.members.length === 1 ? 'member' : 'members'
                        }`}
                    />
                    <div className="grid gap-4 xl:grid-cols-2">
                        <MembersCard team={team} />
                        <Card>
                            <CardHeader>
                                <CardTitle>Grants</CardTitle>
                                <CardDescription>What every member of this team can read and publish.</CardDescription>
                            </CardHeader>
                            <CardContent>
                                <GrantEditor grants={team.grants} target={{ team: team.name }} />
                            </CardContent>
                        </Card>
                    </div>
                    <DeleteTeamCard team={team} />
                </>
            )}
        </>
    );
}
