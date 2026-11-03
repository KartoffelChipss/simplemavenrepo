import { useMe } from '@/api/auth.ts';
import { useBrowseArtifacts } from '@/api/browse.ts';
import type { BrowseArtifact } from '@/api/types.ts';
import { artifactBrowseUrl } from '@/components/artifacts/urls.ts';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Empty, EmptyContent, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty';
import { InputGroup, InputGroupAddon, InputGroupInput } from '@/components/ui/input-group';
import { Skeleton } from '@/components/ui/skeleton';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { formatDate } from '@/lib/format.ts';
import { ChevronRight, Lock, Package, PackageSearch, Search } from 'lucide-react';
import { Link, useLocation, useNavigate, useSearchParams } from 'react-router';

function matches({ artifact }: BrowseArtifact, query: string) {
    const q = query.trim().toLowerCase();
    return (
        !q ||
        `${artifact.groupId}:${artifact.artifactId}`.toLowerCase().includes(q) ||
        (artifact.description?.toLowerCase().includes(q) ?? false)
    );
}

function NoArtifacts() {
    const { data: me } = useMe();
    const location = useLocation();

    return (
        <Empty className="border">
            <EmptyHeader>
                <EmptyMedia variant="icon">
                    <PackageSearch />
                </EmptyMedia>
                <EmptyTitle>Nothing to browse yet</EmptyTitle>
                <EmptyDescription>
                    {me ? 'There are no artifacts you have access to.' : 'There are no public artifacts.'}
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

function ArtifactTable({ items, query }: { items: BrowseArtifact[]; query: string }) {
    const navigate = useNavigate();

    return (
        <div className="rounded-lg border">
            <Table>
                <TableHeader>
                    <TableRow>
                        <TableHead>Artifact</TableHead>
                        <TableHead>Latest</TableHead>
                        <TableHead className="hidden text-right sm:table-cell">Versions</TableHead>
                        <TableHead className="hidden md:table-cell">Updated</TableHead>
                        <TableHead className="w-0">
                            <span className="sr-only">Open</span>
                        </TableHead>
                    </TableRow>
                </TableHeader>
                <TableBody>
                    {items.length === 0 && (
                        <TableRow>
                            <TableCell colSpan={5} className="h-20 text-center text-muted-foreground">
                                No artifact matches “{query.trim()}”.
                            </TableCell>
                        </TableRow>
                    )}
                    {items.map(({ artifact, latestVersion, versionCount, updatedAt }) => {
                        const to = artifactBrowseUrl(artifact);
                        return (
                            <TableRow key={artifact.id} className="cursor-pointer" onClick={() => void navigate(to)}>
                                <TableCell className="max-w-md whitespace-normal">
                                    <div className="flex items-center gap-2">
                                        <Package className="size-4 shrink-0 text-muted-foreground" />
                                        <Link
                                            to={to}
                                            className="font-mono text-sm hover:underline"
                                            onClick={(e) => e.stopPropagation()}
                                        >
                                            <span className="text-muted-foreground">{artifact.groupId}:</span>
                                            <span className="font-medium">{artifact.artifactId}</span>
                                        </Link>
                                        {!artifact.public && (
                                            <Badge variant="outline">
                                                <Lock />
                                                Private
                                            </Badge>
                                        )}
                                    </div>
                                    {artifact.description && (
                                        <p className="mt-0.5 ml-6 line-clamp-1 text-sm text-muted-foreground">
                                            {artifact.description}
                                        </p>
                                    )}
                                </TableCell>
                                <TableCell className="align-top">
                                    {latestVersion ? (
                                        <Badge variant="secondary" className="font-mono">
                                            {latestVersion}
                                        </Badge>
                                    ) : (
                                        <span className="text-sm text-muted-foreground">None yet</span>
                                    )}
                                </TableCell>
                                <TableCell className="hidden text-right align-top text-muted-foreground tabular-nums sm:table-cell">
                                    {versionCount}
                                </TableCell>
                                <TableCell className="hidden align-top text-muted-foreground md:table-cell">
                                    {updatedAt ? formatDate(updatedAt) : ''}
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

export function ArtifactList() {
    const { data: items, error, isPending } = useBrowseArtifacts();
    const [searchParams, setSearchParams] = useSearchParams();
    const query = searchParams.get('q') ?? '';

    const setQuery = (q: string) => setSearchParams(q ? { q } : {}, { replace: true });

    return (
        <>
            {isPending && (
                <div className="flex flex-col gap-2">
                    <Skeleton className="h-9 max-w-sm" />
                    <Skeleton className="h-10" />
                    <Skeleton className="h-10" />
                </div>
            )}
            {error && <p className="text-destructive">{error.message}</p>}
            {items?.length === 0 && <NoArtifacts />}

            {items && items.length > 0 && (
                <>
                    <InputGroup className="max-w-sm">
                        <InputGroupAddon>
                            <Search />
                        </InputGroupAddon>
                        <InputGroupInput
                            type="search"
                            placeholder="Search artifacts…"
                            aria-label="Search artifacts"
                            value={query}
                            onChange={(e) => setQuery(e.target.value)}
                        />
                    </InputGroup>
                    <ArtifactTable items={items.filter((item) => matches(item, query))} query={query} />
                </>
            )}
        </>
    );
}
