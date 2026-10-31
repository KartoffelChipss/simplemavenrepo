import { useMe } from '@/api/auth.ts';
import { useUsers } from '@/api/users.ts';
import { PageHeader } from '@/components/PageHeader.tsx';
import { CreateUserDialog } from '@/components/users/CreateUserDialog.tsx';
import { Badge } from '@/components/ui/badge';
import { InputGroup, InputGroupAddon, InputGroupInput } from '@/components/ui/input-group';
import { Skeleton } from '@/components/ui/skeleton';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { formatDate } from '@/lib/format.ts';
import { ChevronRight, Search } from 'lucide-react';
import { useMemo, useState } from 'react';
import { Link, useNavigate } from 'react-router';

export function UsersPage() {
    const { data: users, error, isPending } = useUsers();
    const { data: me } = useMe();
    const navigate = useNavigate();
    const [search, setSearch] = useState('');

    const filtered = useMemo(() => {
        const q = search.trim().toLowerCase();
        if (!users || !q) return users;
        return users.filter((u) => u.username.toLowerCase().includes(q));
    }, [users, search]);

    return (
        <>
            <div className="flex flex-wrap items-start justify-between gap-4">
                <PageHeader
                    title="Users"
                    description="Everyone who can log in, publish or download private artifacts."
                />
                <CreateUserDialog />
            </div>

            {error && <p className="text-destructive">{error.message}</p>}

            {isPending && (
                <div className="flex flex-col gap-2">
                    <Skeleton className="h-9 max-w-sm" />
                    <Skeleton className="h-10" />
                    <Skeleton className="h-10" />
                </div>
            )}

            {filtered && (
                <>
                    <InputGroup className="max-w-sm">
                        <InputGroupInput
                            aria-label="Search users"
                            placeholder="Search users…"
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
                                    <TableHead>Username</TableHead>
                                    <TableHead>Role</TableHead>
                                    <TableHead>Web login</TableHead>
                                    <TableHead>Created</TableHead>
                                    <TableHead className="w-0">
                                        <span className="sr-only">Open</span>
                                    </TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {filtered.length === 0 && (
                                    <TableRow>
                                        <TableCell colSpan={5} className="h-20 text-center text-muted-foreground">
                                            No users match “{search.trim()}”.
                                        </TableCell>
                                    </TableRow>
                                )}
                                {filtered.map((user) => {
                                    const to = `/users/${encodeURIComponent(user.username)}`;
                                    return (
                                        <TableRow
                                            key={user.id}
                                            className="cursor-pointer"
                                            onClick={() => void navigate(to)}
                                        >
                                            <TableCell className="font-medium">
                                                <Link to={to} onClick={(e) => e.stopPropagation()}>
                                                    {user.username}
                                                </Link>
                                                {user.id === me?.id && (
                                                    <span className="ml-2 text-muted-foreground">(you)</span>
                                                )}
                                            </TableCell>
                                            <TableCell>
                                                {user.isAdmin ? (
                                                    <Badge>Administrator</Badge>
                                                ) : (
                                                    <Badge variant="outline">User</Badge>
                                                )}
                                            </TableCell>
                                            <TableCell>
                                                {user.hasPassword ? (
                                                    'Password'
                                                ) : (
                                                    <span className="text-muted-foreground">Tokens only</span>
                                                )}
                                            </TableCell>
                                            <TableCell>{formatDate(user.createdAt)}</TableCell>
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
