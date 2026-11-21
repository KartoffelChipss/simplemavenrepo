import { useBrowseArtifacts } from '@/api/browse.ts';
import { useRepositories, versionPolicyLabels } from '@/api/repositories.ts';
import { browseUrl } from '@/components/browse/paths.ts';
import { CopyButton } from '@/components/CopyButton.tsx';
import { Badge } from '@/components/ui/badge';
import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty';
import { Skeleton } from '@/components/ui/skeleton';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { ChevronRight, Warehouse } from 'lucide-react';
import { Link, useNavigate } from 'react-router';

export function RepositoryList() {
    const { data: repositories, error, isPending } = useRepositories();
    const { data: artifacts } = useBrowseArtifacts();
    const navigate = useNavigate();

    const visibleCount = (name: string) => artifacts?.filter((a) => a.artifact.repository === name).length;

    if (isPending) {
        return (
            <div className="flex flex-col gap-2">
                <Skeleton className="h-10" />
                <Skeleton className="h-10" />
            </div>
        );
    }
    if (error) return <p className="text-destructive">{error.message}</p>;
    if (repositories.length === 0) {
        return (
            <Empty className="border">
                <EmptyHeader>
                    <EmptyMedia variant="icon">
                        <Warehouse />
                    </EmptyMedia>
                    <EmptyTitle>No repositories</EmptyTitle>
                    <EmptyDescription>An admin has to create a repository first.</EmptyDescription>
                </EmptyHeader>
            </Empty>
        );
    }

    return (
        <div className="rounded-lg border">
            <Table>
                <TableHeader>
                    <TableRow>
                        <TableHead>Repository</TableHead>
                        <TableHead className="hidden sm:table-cell">Versions</TableHead>
                        <TableHead className="text-right">Artifacts</TableHead>
                        <TableHead className="w-0">
                            <span className="sr-only">Open</span>
                        </TableHead>
                    </TableRow>
                </TableHeader>
                <TableBody>
                    {repositories.map((repository) => {
                        const to = browseUrl(repository.name);
                        return (
                            <TableRow key={repository.id} className="cursor-pointer" onClick={() => void navigate(to)}>
                                <TableCell className="max-w-md whitespace-normal">
                                    <div className="flex items-center gap-2">
                                        <Warehouse className="size-4 shrink-0 text-muted-foreground" />
                                        <Link
                                            to={to}
                                            className="font-mono text-sm font-medium hover:underline"
                                            onClick={(e) => e.stopPropagation()}
                                        >
                                            {repository.name}
                                        </Link>
                                        <span onClick={(e) => e.stopPropagation()}>
                                            <CopyButton
                                                text={`${window.location.origin}/${repository.name}`}
                                                label="Copy repository URL"
                                                size="icon-xs"
                                            />
                                        </span>
                                    </div>
                                    {repository.description && (
                                        <p className="mt-0.5 ml-6 line-clamp-2 text-sm text-muted-foreground">
                                            {repository.description}
                                        </p>
                                    )}
                                </TableCell>
                                <TableCell className="hidden align-top sm:table-cell">
                                    <Badge variant="secondary">{versionPolicyLabels[repository.versionPolicy]}</Badge>
                                </TableCell>
                                <TableCell className="text-right align-top text-muted-foreground tabular-nums">
                                    {visibleCount(repository.name)}
                                </TableCell>
                                <TableCell className="align-top">
                                    <ChevronRight className="size-4 text-muted-foreground" />
                                </TableCell>
                            </TableRow>
                        );
                    })}
                </TableBody>
            </Table>
        </div>
    );
}
