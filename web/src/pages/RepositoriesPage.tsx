import { useArtifacts } from '@/api/artifacts.ts';
import { useDeleteRepository, useRepositories, versionPolicyLabels } from '@/api/repositories.ts';
import type { Repository } from '@/api/types.ts';
import { CopyButton } from '@/components/CopyButton.tsx';
import { PageHeader } from '@/components/PageHeader.tsx';
import { RepositoryDialog } from '@/components/repositories/RepositoryDialog.tsx';
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
import { Checkbox } from '@/components/ui/checkbox';
import { Empty, EmptyContent, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty';
import { Field, FieldLabel } from '@/components/ui/field';
import { Skeleton } from '@/components/ui/skeleton';
import { Spinner } from '@/components/ui/spinner';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { toast } from '@/components/ui/toast';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { formatDate } from '@/lib/format.ts';
import { Pencil, Plus, Trash2, Warehouse } from 'lucide-react';
import { useState } from 'react';

const createButton = (
    <Button>
        <Plus />
        Create repository
    </Button>
);

const repositoryUrl = (name: string) => `${window.location.origin}/${name}`;

function DeleteRepositoryButton({ repository, artifactCount }: { repository: Repository; artifactCount: number }) {
    const remove = useDeleteRepository();
    const [deleteFiles, setDeleteFiles] = useState(false);
    const button = (
        <Button variant="ghost" size="icon-sm" aria-label={`Delete ${repository.name}`} disabled={artifactCount > 0}>
            <Trash2 />
        </Button>
    );

    if (artifactCount > 0) {
        return (
            <Tooltip>
                {/* Disabled buttons don't get pointer events, so the span shows the tooltip */}
                <TooltipTrigger render={<span />}>{button}</TooltipTrigger>
                <TooltipContent>
                    Delete its {artifactCount} {artifactCount === 1 ? 'artifact' : 'artifacts'} first
                </TooltipContent>
            </Tooltip>
        );
    }

    function onDelete() {
        remove.mutate(
            { name: repository.name, deleteFiles },
            {
                onSuccess: () => toast.add({ title: 'Repository deleted', description: repository.name }),
                onError: (err) => toast.add({ title: 'Could not delete repository', description: err.message }),
            }
        );
    }

    return (
        <AlertDialog>
            <AlertDialogTrigger render={button} />
            <AlertDialogContent>
                <AlertDialogHeader>
                    <AlertDialogTitle>Delete {repository.name}?</AlertDialogTitle>
                    <AlertDialogDescription>
                        Builds can no longer fetch from or deploy to {repositoryUrl(repository.name)}. Grants and token
                        scopes for it are kept and apply again if a repository of the same name is created.
                    </AlertDialogDescription>
                </AlertDialogHeader>
                <Field orientation="horizontal">
                    <Checkbox id="repository-delete-files" checked={deleteFiles} onCheckedChange={setDeleteFiles} />
                    <FieldLabel htmlFor="repository-delete-files">
                        Also delete files left over from unregistered artifacts
                    </FieldLabel>
                </Field>
                <AlertDialogFooter>
                    <AlertDialogCancel>Cancel</AlertDialogCancel>
                    <AlertDialogAction variant="destructive" disabled={remove.isPending} onClick={onDelete}>
                        {remove.isPending && <Spinner />}
                        Delete
                    </AlertDialogAction>
                </AlertDialogFooter>
            </AlertDialogContent>
        </AlertDialog>
    );
}

export function RepositoriesPage() {
    const { data: repositories, error, isPending } = useRepositories();
    const { data: artifacts } = useArtifacts();

    const artifactCount = (name: string) => artifacts?.filter((a) => a.repository === name).length ?? 0;

    return (
        <>
            <div className="flex flex-wrap items-start justify-between gap-4">
                <PageHeader
                    title="Repositories"
                    description="Each repository is served under its own path and decides which versions it accepts."
                />
                <RepositoryDialog trigger={createButton} />
            </div>

            {error && <p className="text-destructive">{error.message}</p>}

            {isPending && (
                <div className="flex flex-col gap-2">
                    <Skeleton className="h-10" />
                    <Skeleton className="h-10" />
                </div>
            )}

            {repositories?.length === 0 && (
                <Empty className="border">
                    <EmptyHeader>
                        <EmptyMedia variant="icon">
                            <Warehouse />
                        </EmptyMedia>
                        <EmptyTitle>No repositories</EmptyTitle>
                        <EmptyDescription>Create a repository to register artifacts in it.</EmptyDescription>
                    </EmptyHeader>
                    <EmptyContent>
                        <RepositoryDialog trigger={createButton} />
                    </EmptyContent>
                </Empty>
            )}

            {repositories && repositories.length > 0 && (
                <div className="rounded-lg border">
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead>Repository</TableHead>
                                <TableHead>Versions</TableHead>
                                <TableHead className="hidden text-right sm:table-cell">Artifacts</TableHead>
                                <TableHead className="hidden md:table-cell">Created</TableHead>
                                <TableHead className="w-0">
                                    <span className="sr-only">Actions</span>
                                </TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {repositories.map((repository) => (
                                <TableRow key={repository.id}>
                                    <TableCell className="max-w-md whitespace-normal">
                                        <div className="flex items-center gap-1">
                                            <span className="font-mono text-sm font-medium">{repository.name}</span>
                                            <CopyButton
                                                text={repositoryUrl(repository.name)}
                                                label="Copy URL"
                                                variant="ghost"
                                                size="icon-xs"
                                            />
                                        </div>
                                        {repository.description && (
                                            <p className="mt-0.5 line-clamp-2 text-sm text-muted-foreground">
                                                {repository.description}
                                            </p>
                                        )}
                                    </TableCell>
                                    <TableCell>
                                        <Badge variant="secondary">
                                            {versionPolicyLabels[repository.versionPolicy]}
                                        </Badge>
                                    </TableCell>
                                    <TableCell className="hidden text-right tabular-nums sm:table-cell">
                                        {artifacts ? artifactCount(repository.name) : ''}
                                    </TableCell>
                                    <TableCell className="hidden md:table-cell">
                                        {formatDate(repository.createdAt)}
                                    </TableCell>
                                    <TableCell>
                                        <div className="flex justify-end gap-1">
                                            <RepositoryDialog
                                                repository={repository}
                                                trigger={
                                                    <Button
                                                        variant="ghost"
                                                        size="icon-sm"
                                                        aria-label={`Edit ${repository.name}`}
                                                    >
                                                        <Pencil />
                                                    </Button>
                                                }
                                            />
                                            <DeleteRepositoryButton
                                                repository={repository}
                                                artifactCount={artifactCount(repository.name)}
                                            />
                                        </div>
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
