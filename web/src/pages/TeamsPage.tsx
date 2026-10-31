import { useTeams } from '@/api/teams.ts';
import { PageHeader } from '@/components/PageHeader.tsx';
import { CreateTeamDialog } from '@/components/teams/CreateTeamDialog.tsx';
import { Empty, EmptyContent, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty';
import { InputGroup, InputGroupAddon, InputGroupInput } from '@/components/ui/input-group';
import { Skeleton } from '@/components/ui/skeleton';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { formatDate } from '@/lib/format.ts';
import { ChevronRight, Search, UserGroup } from 'lucide-react';
import { useMemo, useState } from 'react';
import { Link, useNavigate } from 'react-router';

export function TeamsPage() {
    const { data: teams, error, isPending } = useTeams();
    const navigate = useNavigate();
    const [search, setSearch] = useState('');

    const filtered = useMemo(() => {
        const q = search.trim().toLowerCase();
        if (!teams || !q) return teams;
        return teams.filter((t) => t.name.toLowerCase().includes(q));
    }, [teams, search]);

    return (
        <>
            <div className="flex flex-wrap items-start justify-between gap-4">
                <PageHeader title="Teams" description="Groups of users that share grants." />
                <CreateTeamDialog />
            </div>

            {error && <p className="text-destructive">{error.message}</p>}

            {isPending && (
                <div className="flex flex-col gap-2">
                    <Skeleton className="h-9 max-w-sm" />
                    <Skeleton className="h-10" />
                    <Skeleton className="h-10" />
                </div>
            )}

            {teams?.length === 0 && (
                <Empty className="border">
                    <EmptyHeader>
                        <EmptyMedia variant="icon">
                            <UserGroup />
                        </EmptyMedia>
                        <EmptyTitle>No teams yet</EmptyTitle>
                        <EmptyDescription>
                            Create a team to give several users the same access at once.
                        </EmptyDescription>
                    </EmptyHeader>
                    <EmptyContent>
                        <CreateTeamDialog />
                    </EmptyContent>
                </Empty>
            )}

            {teams && teams.length > 0 && filtered && (
                <>
                    <InputGroup className="max-w-sm">
                        <InputGroupInput
                            aria-label="Search teams"
                            placeholder="Search teams…"
                            value={search}
                            onChange={(e) => setSearch(e.target.value)}
                        />
                        <InputGroupAddon>
                            <Search />
                        </InputGroupAddon>
                    </InputGroup>
                    <div className="rounded-lg border">
                        <Table>
                            <TableHeader>
                                <TableRow>
                                    <TableHead>Name</TableHead>
                                    <TableHead>Created</TableHead>
                                    <TableHead className="w-0">
                                        <span className="sr-only">Open</span>
                                    </TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {filtered.length === 0 && (
                                    <TableRow>
                                        <TableCell colSpan={3} className="h-20 text-center text-muted-foreground">
                                            No teams match “{search.trim()}”.
                                        </TableCell>
                                    </TableRow>
                                )}
                                {filtered.map((team) => {
                                    const to = `/teams/${encodeURIComponent(team.name)}`;
                                    return (
                                        // The whole row opens the team; the link inside keeps it reachable by keyboard.
                                        <TableRow
                                            key={team.id}
                                            className="cursor-pointer"
                                            onClick={() => void navigate(to)}
                                        >
                                            <TableCell className="font-medium">
                                                <Link to={to} onClick={(e) => e.stopPropagation()}>
                                                    {team.name}
                                                </Link>
                                            </TableCell>
                                            <TableCell>{formatDate(team.createdAt)}</TableCell>
                                            <TableCell>
                                                <ChevronRight className="size-4 text-muted-foreground" />
                                            </TableCell>
                                        </TableRow>
                                    );
                                })}
                            </TableBody>
                        </Table>
                    </div>
                </>
            )}
        </>
    );
}
