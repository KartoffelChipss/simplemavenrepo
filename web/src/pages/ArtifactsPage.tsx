import { useArtifacts } from '@/api/artifacts.ts';
import type { Artifact } from '@/api/types.ts';
import { ArtifactDialog } from '@/components/artifacts/ArtifactDialog.tsx';
import { artifactSettingsUrl } from '@/components/artifacts/urls.ts';
import { PageHeader } from '@/components/PageHeader.tsx';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Empty, EmptyContent, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty';
import { InputGroup, InputGroupAddon, InputGroupInput } from '@/components/ui/input-group';
import { Skeleton } from '@/components/ui/skeleton';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { formatDate } from '@/lib/format.ts';
import { Boxes, ChevronRight, Globe, Lock, Plus, Search } from 'lucide-react';
import { useMemo, useState } from 'react';
import { Link, useNavigate } from 'react-router';

const coordinates = (a: Artifact) => `${a.groupId}:${a.artifactId}`;

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
    const navigate = useNavigate();

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
                                        <span className="sr-only">Open</span>
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
                                {filtered.map((artifact) => {
                                    const to = artifactSettingsUrl(artifact);
                                    return (
                                        <TableRow
                                            key={artifact.id}
                                            className="cursor-pointer"
                                            onClick={() => void navigate(to)}
                                        >
                                            <TableCell className="max-w-md whitespace-normal">
                                                <Link
                                                    to={to}
                                                    className="font-mono text-sm"
                                                    onClick={(e) => e.stopPropagation()}
                                                >
                                                    <span className="text-muted-foreground">{artifact.groupId}:</span>
                                                    <span className="font-medium">{artifact.artifactId}</span>
                                                </Link>
                                                {artifact.description && (
                                                    <p className="mt-0.5 line-clamp-1 text-sm text-muted-foreground">
                                                        {artifact.description}
                                                    </p>
                                                )}
                                            </TableCell>
                                            <TableCell>
                                                <Visibility isPublic={artifact.public} />
                                            </TableCell>
                                            <TableCell>{formatDate(artifact.createdAt)}</TableCell>
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
