import { useMe } from '@/api/auth.ts';
import { javadocPageUrl, useJavadocList } from '@/api/javadoc.ts';
import type { JavadocArtifact } from '@/api/types.ts';
import { PageHeader } from '@/components/PageHeader.tsx';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Empty, EmptyContent, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty';
import { InputGroup, InputGroupAddon, InputGroupInput } from '@/components/ui/input-group';
import { Skeleton } from '@/components/ui/skeleton';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { BookOpen, BookX, Lock, Search } from 'lucide-react';
import { useState } from 'react';
import { Link, useLocation, useNavigate } from 'react-router';

/** How many older versions a row links to before summarizing the rest. */
const OLDER_VERSIONS = 4;

const docsUrl = (item: JavadocArtifact, version: string) => javadocPageUrl(`${item.artifact.path}/${version}`);

function matches(item: JavadocArtifact, query: string) {
    const q = query.trim().toLowerCase();
    const { groupId, artifactId, description } = item.artifact;
    return (
        !q || `${groupId}:${artifactId}`.toLowerCase().includes(q) || (description?.toLowerCase().includes(q) ?? false)
    );
}

function NoJavadoc() {
    const { data: me } = useMe();
    const location = useLocation();

    return (
        <Empty className="border">
            <EmptyHeader>
                <EmptyMedia variant="icon">
                    <BookX />
                </EmptyMedia>
                <EmptyTitle>No javadoc yet</EmptyTitle>
                <EmptyDescription>
                    {me
                        ? 'None of the artifacts you have access to has a javadoc jar.'
                        : 'None of the public artifacts has a javadoc jar.'}{' '}
                    Publish one with the classifier <code className="font-mono">javadoc</code> to browse it here.
                </EmptyDescription>
            </EmptyHeader>
            {!me && (
                <EmptyContent>
                    <Button render={<Link to="/login" state={{ from: location.pathname }} />} nativeButton={false}>
                        Log in to see private artifacts
                    </Button>
                </EmptyContent>
            )}
        </Empty>
    );
}

function JavadocTable({ items }: { items: JavadocArtifact[] }) {
    const navigate = useNavigate();

    return (
        <div className="rounded-lg border">
            <Table>
                <TableHeader>
                    <TableRow>
                        <TableHead>Artifact</TableHead>
                        <TableHead>Latest</TableHead>
                        <TableHead className="hidden md:table-cell">Older versions</TableHead>
                    </TableRow>
                </TableHeader>
                <TableBody>
                    {items.map((item) => {
                        const { artifact, versions } = item;
                        const [latest, ...older] = versions;
                        return (
                            <TableRow
                                key={artifact.id}
                                className="cursor-pointer"
                                onClick={() => void navigate(docsUrl(item, latest))}
                            >
                                <TableCell className="whitespace-normal">
                                    <div className="flex items-center gap-2">
                                        <BookOpen className="size-4 shrink-0 text-muted-foreground" />
                                        <Link
                                            to={docsUrl(item, latest)}
                                            className="font-mono text-sm hover:underline"
                                            onClick={(e) => e.stopPropagation()}
                                        >
                                            <span className="text-muted-foreground">{artifact.groupId}:</span>
                                            {artifact.artifactId}
                                        </Link>
                                        {!artifact.public && (
                                            <Badge variant="outline">
                                                <Lock />
                                                Private
                                            </Badge>
                                        )}
                                    </div>
                                    {artifact.description && (
                                        <p className="mt-1 ml-6 text-sm text-muted-foreground">
                                            {artifact.description}
                                        </p>
                                    )}
                                </TableCell>
                                <TableCell className="align-top">
                                    <Badge variant="secondary" className="font-mono">
                                        {latest}
                                    </Badge>
                                </TableCell>
                                <TableCell className="hidden align-top whitespace-normal md:table-cell">
                                    <div className="flex flex-wrap gap-x-3 gap-y-1 font-mono text-sm">
                                        {older.slice(0, OLDER_VERSIONS).map((v) => (
                                            <Link
                                                key={v}
                                                to={docsUrl(item, v)}
                                                className="text-muted-foreground hover:text-foreground hover:underline"
                                                onClick={(e) => e.stopPropagation()}
                                            >
                                                {v}
                                            </Link>
                                        ))}
                                        {older.length > OLDER_VERSIONS && (
                                            <span className="font-sans text-muted-foreground">
                                                +{older.length - OLDER_VERSIONS} more
                                            </span>
                                        )}
                                    </div>
                                </TableCell>
                            </TableRow>
                        );
                    })}
                </TableBody>
            </Table>
        </div>
    );
}

export function JavadocIndexPage() {
    const { data: items, error, isPending } = useJavadocList();
    const [query, setQuery] = useState('');
    const filtered = items?.filter((item) => matches(item, query)) ?? [];

    return (
        <>
            <PageHeader title="Javadoc" description="API documentation of the artifacts you have access to." />

            {isPending && (
                <div className="flex flex-col gap-2">
                    <Skeleton className="h-9 max-w-sm" />
                    <Skeleton className="h-10" />
                    <Skeleton className="h-10" />
                </div>
            )}
            {error && <p className="text-destructive">{error.message}</p>}
            {items?.length === 0 && <NoJavadoc />}

            {items && items.length > 0 && (
                <>
                    <InputGroup className="max-w-sm">
                        <InputGroupAddon>
                            <Search />
                        </InputGroupAddon>
                        <InputGroupInput
                            type="search"
                            placeholder="Filter artifacts"
                            aria-label="Filter artifacts"
                            value={query}
                            onChange={(e) => setQuery(e.target.value)}
                        />
                    </InputGroup>
                    {filtered.length > 0 ? (
                        <JavadocTable items={filtered} />
                    ) : (
                        <p className="text-sm text-muted-foreground">No artifact matches “{query.trim()}”.</p>
                    )}
                </>
            )}
        </>
    );
}
