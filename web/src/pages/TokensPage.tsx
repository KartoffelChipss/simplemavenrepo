import { useDeleteMyToken, useMyTokens } from '@/api/tokens.ts';
import type { Token } from '@/api/types.ts';
import { PageHeader } from '@/components/PageHeader.tsx';
import { CreateTokenDialog } from '@/components/tokens/CreateTokenDialog.tsx';
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
import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty';
import { Skeleton } from '@/components/ui/skeleton';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { toast } from '@/components/ui/toast';
import { formatDate, formatDateTime } from '@/lib/format.ts';
import { KeyRound, Trash2 } from 'lucide-react';

function Expiry({ expiresAt }: { expiresAt: string | null }) {
    if (!expiresAt) return <span className="text-muted-foreground">Never</span>;
    if (new Date(expiresAt) <= new Date()) return <Badge variant="destructive">Expired</Badge>;
    return <>{formatDate(expiresAt)}</>;
}

function RevokeButton({ token }: { token: Token }) {
    const remove = useDeleteMyToken();

    function revoke() {
        remove.mutate(token.id, {
            onSuccess: () => toast.add({ title: 'Token revoked', description: token.name }),
            onError: (err) => toast.add({ title: 'Could not revoke token', description: err.message }),
        });
    }

    return (
        <AlertDialog>
            <AlertDialogTrigger render={<Button variant="ghost" size="icon-sm" aria-label={`Revoke ${token.name}`} />}>
                <Trash2 />
            </AlertDialogTrigger>
            <AlertDialogContent>
                <AlertDialogHeader>
                    <AlertDialogTitle>Revoke {token.name}?</AlertDialogTitle>
                    <AlertDialogDescription>
                        Builds and scripts using this token will no longer be able to authenticate. This can't be
                        undone.
                    </AlertDialogDescription>
                </AlertDialogHeader>
                <AlertDialogFooter>
                    <AlertDialogCancel>Cancel</AlertDialogCancel>
                    <AlertDialogAction variant="destructive" disabled={remove.isPending} onClick={revoke}>
                        Revoke
                    </AlertDialogAction>
                </AlertDialogFooter>
            </AlertDialogContent>
        </AlertDialog>
    );
}

export function TokensPage() {
    const { data: tokens, error, isPending } = useMyTokens();

    return (
        <>
            <div className="flex flex-wrap items-start justify-between gap-4">
                <PageHeader
                    title="Tokens"
                    description="Tokens let Maven, Gradle and scripts access the repository as you."
                />
                <CreateTokenDialog />
            </div>

            {error && <p className="text-destructive">{error.message}</p>}

            {isPending && (
                <div className="flex flex-col gap-2">
                    <Skeleton className="h-10" />
                    <Skeleton className="h-10" />
                </div>
            )}

            {tokens?.length === 0 && (
                <Empty className="border">
                    <EmptyHeader>
                        <EmptyMedia variant="icon">
                            <KeyRound />
                        </EmptyMedia>
                        <EmptyTitle>No tokens yet</EmptyTitle>
                        <EmptyDescription>
                            Create a token to publish or download artifacts with your build tool.
                        </EmptyDescription>
                    </EmptyHeader>
                </Empty>
            )}

            {tokens && tokens.length > 0 && (
                <div className="rounded-lg border">
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead>Name</TableHead>
                                <TableHead>Token</TableHead>
                                <TableHead>Created</TableHead>
                                <TableHead>Last used</TableHead>
                                <TableHead>Expires</TableHead>
                                <TableHead className="w-0">
                                    <span className="sr-only">Actions</span>
                                </TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {tokens.map((token) => (
                                <TableRow key={token.id}>
                                    <TableCell className="font-medium">{token.name}</TableCell>
                                    <TableCell className="font-mono text-muted-foreground">{token.prefix}…</TableCell>
                                    <TableCell>{formatDate(token.createdAt)}</TableCell>
                                    <TableCell>
                                        {token.lastUsedAt ? (
                                            formatDateTime(token.lastUsedAt)
                                        ) : (
                                            <span className="text-muted-foreground">Never</span>
                                        )}
                                    </TableCell>
                                    <TableCell>
                                        <Expiry expiresAt={token.expiresAt} />
                                    </TableCell>
                                    <TableCell>
                                        <RevokeButton token={token} />
                                    </TableCell>
                                </TableRow>
                            ))}
                        </TableBody>
                    </Table>
                </div>
            )}
        </>
    );
}
