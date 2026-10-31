import { useArtifacts, useDeleteArtifact } from '@/api/artifacts.ts';
import type { Artifact } from '@/api/types.ts';
import { ArtifactDialog } from '@/components/artifacts/ArtifactDialog.tsx';
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
import { Empty, EmptyContent, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty';
import { InputGroup, InputGroupAddon, InputGroupInput } from '@/components/ui/input-group';
import { Skeleton } from '@/components/ui/skeleton';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { toast } from '@/components/ui/toast';
import { formatDate } from '@/lib/format.ts';
import { Boxes, Globe, Lock, Pencil, Plus, Search, Trash2 } from 'lucide-react';
import { useMemo, useState } from 'react';

const coordinates = (a: Artifact) => `${a.groupId}:${a.artifactId}`;

function DeleteButton({ artifact }: { artifact: Artifact }) {
    const remove = useDeleteArtifact();

    function onDelete() {
        remove.mutate(artifact, {
            onSuccess: () => toast.add({ title: 'Artifact deleted', description: coordinates(artifact) }),
            onError: (err) => toast.add({ title: 'Could not delete artifact', description: err.message }),
        });
    }

    return (
        <AlertDialog>
            <AlertDialogTrigger
                render={<Button variant="ghost" size="icon-sm" aria-label={`Delete ${coordinates(artifact)}`} />}
            >
                <Trash2 />
            </AlertDialogTrigger>
            <AlertDialogContent>
                <AlertDialogHeader>
                    <AlertDialogTitle>Delete {coordinates(artifact)}?</AlertDialogTitle>
                    <AlertDialogDescription>
                        It can no longer be published or downloaded. Files already uploaded stay on disk and become
                        available again if you register the artifact again.
                    </AlertDialogDescription>
                </AlertDialogHeader>
                <AlertDialogFooter>
                    <AlertDialogCancel>Cancel</AlertDialogCancel>
                    <AlertDialogAction variant="destructive" disabled={remove.isPending} onClick={onDelete}>
                        Delete
                    </AlertDialogAction>
                </AlertDialogFooter>
            </AlertDialogContent>
        </AlertDialog>
    );
}

function Visibility({ isPublic }: { isPublic: boolean }) {
    return isPublic ? (
        <Badge variant="secondary">
            <Globe />
            Public
        </Badge>
    ) : (
        <Badge variant="outline">
            <Lock />
            Private
        </Badge>
    );
}

const registerButton = (
    <Button>
        <Plus />
        Register artifact
    </Button>
);

export function ArtifactsPage() {
    const { data: artifacts, error, isPending } = useArtifacts();
    const [search, setSearch] = useState('');

    const filtered = useMemo(() => {
        const q = search.trim().toLowerCase();
        if (!artifacts || !q) return artifacts;
        return artifacts.filter((a) =>
            [coordinates(a), a.description ?? ''].some((text) => text.toLowerCase().includes(q))
        );
    }, [artifacts, search]);

    return (
        <>
            <div className="flex flex-wrap items-start justify-between gap-4">
                <PageHeader
                    title="Artifacts"
                    description="Registered artifacts can be published and downloaded. Everything else is rejected."
                />
                <ArtifactDialog trigger={registerButton} />
            </div>

            {error && <p className="text-destructive">{error.message}</p>}

            {isPending && (
                <div className="flex flex-col gap-2">
                    <Skeleton className="h-9 max-w-sm" />
                    <Skeleton className="h-10" />
                    <Skeleton className="h-10" />
                </div>
            )}

            {artifacts?.length === 0 && (
                <Empty className="border">
                    <EmptyHeader>
                        <EmptyMedia variant="icon">
                            <Boxes />
                        </EmptyMedia>
                        <EmptyTitle>No artifacts yet</EmptyTitle>
                        <EmptyDescription>
                            Register an artifact before publishing it with Maven or Gradle.
                        </EmptyDescription>
                    </EmptyHeader>
                    <EmptyContent>
                        <ArtifactDialog trigger={registerButton} />
                    </EmptyContent>
                </Empty>
            )}

            {artifacts && artifacts.length > 0 && filtered && (
                <>
                    <InputGroup className="max-w-sm">
                        <InputGroupInput
                            aria-label="Search artifacts"
                            placeholder="Search artifacts…"
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
                                    <TableHead>Artifact</TableHead>
                                    <TableHead>Visibility</TableHead>
                                    <TableHead>Registered</TableHead>
                                    <TableHead className="w-0">
                                        <span className="sr-only">Actions</span>
                                    </TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {filtered.length === 0 && (
                                    <TableRow>
                                        <TableCell colSpan={4} className="h-20 text-center text-muted-foreground">
                                            No artifacts match “{search.trim()}”.
                                        </TableCell>
                                    </TableRow>
                                )}
                                {filtered.map((artifact) => (
                                    <TableRow key={artifact.id}>
                                        <TableCell className="max-w-md whitespace-normal">
                                            <div className="font-mono text-sm">
                                                <span className="text-muted-foreground">{artifact.groupId}:</span>
                                                <span className="font-medium">{artifact.artifactId}</span>
                                            </div>
                                        </TableCell>
                                        <TableCell>
                                            <Visibility isPublic={artifact.public} />
                                        </TableCell>
                                        <TableCell>{formatDate(artifact.createdAt)}</TableCell>
                                        <TableCell>
                                            <div className="flex">
                                                <ArtifactDialog
                                                    artifact={artifact}
                                                    trigger={
                                                        <Button
                                                            variant="ghost"
                                                            size="icon-sm"
                                                            aria-label={`Edit ${coordinates(artifact)}`}
                                                        >
                                                            <Pencil />
                                                        </Button>
                                                    }
                                                />
                                                <DeleteButton artifact={artifact} />
                                            </div>
                                        </TableCell>
                                    </TableRow>
                                ))}
                            </TableBody>
                        </Table>
                    </div>
                </>
            )}
        </>
    );
}
