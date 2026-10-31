import { useMe } from '@/api/auth.ts';
import { repositoryFileUrl, useBrowse } from '@/api/browse.ts';
import { isApiError } from '@/api/client.ts';
import { javadocPageUrl } from '@/api/javadoc.ts';
import type { BrowseEntry, BrowseListing } from '@/api/types.ts';
import { ArtifactInfo } from '@/components/browse/ArtifactInfo.tsx';
import { PageHeader } from '@/components/PageHeader.tsx';
import { Badge } from '@/components/ui/badge';
import {
    Breadcrumb,
    BreadcrumbItem,
    BreadcrumbLink,
    BreadcrumbList,
    BreadcrumbPage,
    BreadcrumbSeparator,
} from '@/components/ui/breadcrumb';
import { Button } from '@/components/ui/button';
import { Empty, EmptyContent, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty';
import { Skeleton } from '@/components/ui/skeleton';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { formatBytes, formatDateTime } from '@/lib/format.ts';
import { cn } from '@/lib/utils';
import { usePageTitle } from '@/title.ts';
import { BookOpen, Download, File, Folder, FolderSearch, Lock, Package } from 'lucide-react';
import { Link, useLocation, useNavigate, useParams } from 'react-router';

const SIDECAR_EXTENSIONS = ['md5', 'sha1', 'sha256', 'sha512', 'asc'];

const isJavadocJar = (entry: BrowseEntry) => entry.type === 'file' && entry.name.endsWith('-javadoc.jar');

const join = (...parts: string[]) => parts.filter(Boolean).join('/');
const browseUrl = (path: string) => `/browse/${path.split('/').filter(Boolean).map(encodeURIComponent).join('/')}`;

interface Row {
    entry: BrowseEntry;
    sidecars: string[];
}

function groupSidecars(entries: BrowseEntry[]): Row[] {
    const files = new Set(entries.filter((e) => e.type === 'file').map((e) => e.name));
    const isSidecar = (name: string) =>
        SIDECAR_EXTENSIONS.some((ext) => name.endsWith(`.${ext}`) && files.has(name.slice(0, -ext.length - 1)));
    return entries
        .filter((e) => e.type === 'directory' || !isSidecar(e.name))
        .map((entry) => ({
            entry,
            sidecars:
                entry.type === 'file' ? SIDECAR_EXTENSIONS.filter((ext) => files.has(`${entry.name}.${ext}`)) : [],
        }));
}

function PathBreadcrumbs({ path }: { path: string }) {
    const segments = path.split('/').filter(Boolean);
    return (
        <Breadcrumb>
            <BreadcrumbList>
                <BreadcrumbItem>
                    {segments.length === 0 ? (
                        <BreadcrumbPage>Repository</BreadcrumbPage>
                    ) : (
                        <BreadcrumbLink render={<Link to="/browse" />}>Repository</BreadcrumbLink>
                    )}
                </BreadcrumbItem>
                {segments.map((segment, i) => {
                    const last = i === segments.length - 1;
                    return [
                        <BreadcrumbSeparator key={`sep-${i}`} />,
                        <BreadcrumbItem key={i}>
                            {last ? (
                                <BreadcrumbPage className="font-mono">{segment}</BreadcrumbPage>
                            ) : (
                                <BreadcrumbLink
                                    className="font-mono"
                                    render={<Link to={browseUrl(segments.slice(0, i + 1).join('/'))} />}
                                >
                                    {segment}
                                </BreadcrumbLink>
                            )}
                        </BreadcrumbItem>,
                    ];
                })}
            </BreadcrumbList>
        </Breadcrumb>
    );
}

function EntryIcon({ entry }: { entry: BrowseEntry }) {
    const className = 'size-4 shrink-0 text-muted-foreground';
    if (entry.artifact) return <Package className={className} />;
    if (entry.type === 'directory') return <Folder className={className} />;
    return <File className={className} />;
}

function EntryTable({ listing }: { listing: BrowseListing }) {
    const navigate = useNavigate();
    const rows = groupSidecars(listing.entries);

    if (rows.length === 0) {
        return <p className="text-sm text-muted-foreground">This directory is empty.</p>;
    }

    return (
        <div className="rounded-lg border">
            <Table>
                <TableHeader>
                    <TableRow>
                        <TableHead>Name</TableHead>
                        <TableHead className="text-right">Size</TableHead>
                        <TableHead className="hidden sm:table-cell">Modified</TableHead>
                    </TableRow>
                </TableHeader>
                <TableBody>
                    {rows.map(({ entry, sidecars }) => {
                        const path = join(listing.path, entry.name);
                        const isDir = entry.type === 'directory';
                        return (
                            <TableRow
                                key={entry.name}
                                className={cn(isDir && 'cursor-pointer')}
                                onClick={isDir ? () => void navigate(browseUrl(path)) : undefined}
                            >
                                <TableCell className="whitespace-normal">
                                    <div className="flex items-center gap-2">
                                        <EntryIcon entry={entry} />
                                        {isDir ? (
                                            <Link
                                                to={browseUrl(path)}
                                                className="font-mono text-sm hover:underline"
                                                onClick={(e) => e.stopPropagation()}
                                            >
                                                {entry.name}
                                            </Link>
                                        ) : (
                                            <a
                                                href={repositoryFileUrl(path)}
                                                download
                                                className="flex items-center gap-1 font-mono text-sm hover:underline"
                                            >
                                                {entry.name}
                                                <Download className="size-3.5 text-muted-foreground" />
                                            </a>
                                        )}
                                        {isJavadocJar(entry) && (
                                            <Link
                                                to={javadocPageUrl(listing.path)}
                                                className="flex items-center gap-1 rounded border px-1 text-[0.7rem] text-muted-foreground hover:text-foreground"
                                            >
                                                <BookOpen className="size-3" />
                                                View
                                            </Link>
                                        )}
                                        {entry.artifact && !entry.artifact.public && (
                                            <Badge variant="outline">
                                                <Lock />
                                                Private
                                            </Badge>
                                        )}
                                        {sidecars.length > 0 && (
                                            <span className="flex gap-1">
                                                {sidecars.map((ext) => (
                                                    <a
                                                        key={ext}
                                                        href={repositoryFileUrl(`${path}.${ext}`)}
                                                        download
                                                        className="rounded border px-1 font-mono text-[0.7rem] text-muted-foreground hover:text-foreground"
                                                    >
                                                        {ext}
                                                    </a>
                                                ))}
                                            </span>
                                        )}
                                    </div>
                                    {entry.artifact?.description && (
                                        <p className="mt-1 ml-6 text-sm text-muted-foreground">
                                            {entry.artifact.description}
                                        </p>
                                    )}
                                </TableCell>
                                <TableCell className="text-right text-muted-foreground tabular-nums">
                                    {entry.size !== undefined ? formatBytes(entry.size) : ''}
                                </TableCell>
                                <TableCell className="hidden text-muted-foreground sm:table-cell">
                                    {entry.modifiedAt ? formatDateTime(entry.modifiedAt) : ''}
                                </TableCell>
                            </TableRow>
                        );
                    })}
                </TableBody>
            </Table>
        </div>
    );
}

function artifactVersion(listing: BrowseListing): string | undefined {
    const artifact = listing.artifact;
    if (!artifact) return;
    const rest = listing.path.slice(artifact.path.length).split('/').filter(Boolean);
    if (rest.length > 0) return rest[0];
    return listing.entries.find((e) => e.type === 'directory')?.name;
}

function javadocUrl(listing: BrowseListing): string | undefined {
    const artifact = listing.artifact;
    if (!artifact || listing.path.split('/').length !== artifact.path.split('/').length + 1) return;
    return listing.entries.some(isJavadocJar) ? javadocPageUrl(listing.path) : undefined;
}

function NotFound({ path }: { path: string }) {
    const { data: me } = useMe();
    const location = useLocation();
    const root = path === '';

    return (
        <Empty className="border">
            <EmptyHeader>
                <EmptyMedia variant="icon">
                    <FolderSearch />
                </EmptyMedia>
                <EmptyTitle>{root ? 'Nothing to browse yet' : 'Not found'}</EmptyTitle>
                <EmptyDescription>
                    {root
                        ? me
                            ? 'There are no artifacts you have access to.'
                            : 'There are no public artifacts.'
                        : me
                          ? 'This path doesn’t exist or you don’t have access to it.'
                          : 'This path doesn’t exist, or it is private.'}
                </EmptyDescription>
            </EmptyHeader>
            {!me && (
                <EmptyContent>
                    <Button render={<Link to="/login" state={{ from: location.pathname }} />}>
                        Log in to see private artifacts
                    </Button>
                </EmptyContent>
            )}
        </Empty>
    );
}

export function BrowsePage() {
    const path = (useParams()['*'] ?? '').split('/').filter(Boolean).join('/');
    const { data: listing, error, isPending, isPlaceholderData } = useBrowse(path);
    usePageTitle(path.split('/').at(-1) || undefined);

    return (
        <>
            <PageHeader title="Browse" description="The repository as it is laid out on disk." />
            <PathBreadcrumbs path={path} />

            {isPending && (
                <div className="flex flex-col gap-2">
                    <Skeleton className="h-10" />
                    <Skeleton className="h-10" />
                    <Skeleton className="h-10" />
                </div>
            )}

            {isApiError(error, 404) && <NotFound path={path} />}
            {error && !isApiError(error, 404) && <p className="text-destructive">{error.message}</p>}

            {listing && !error && (
                <div className={cn('flex flex-col gap-4 transition-opacity', isPlaceholderData && 'opacity-60')}>
                    {listing.artifact && (
                        <ArtifactInfo
                            artifact={listing.artifact}
                            version={artifactVersion(listing)}
                            javadocUrl={javadocUrl(listing)}
                        />
                    )}
                    <EntryTable listing={listing} />
                </div>
            )}
        </>
    );
}
