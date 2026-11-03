import { useMe } from '@/api/auth.ts';
import { useBrowse } from '@/api/browse.ts';
import { isApiError } from '@/api/client.ts';
import type { Artifact } from '@/api/types.ts';
import { ArtifactDetail } from '@/components/browse/ArtifactDetail.tsx';
import { ArtifactList } from '@/components/browse/ArtifactList.tsx';
import { EntryTable } from '@/components/browse/EntryTable.tsx';
import { browseUrl, join } from '@/components/browse/paths.ts';
import { PageHeader } from '@/components/PageHeader.tsx';
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
import { cn } from '@/lib/utils';
import { usePageTitle } from '@/title.ts';
import { FolderSearch, Search } from 'lucide-react';
import { Link, useLocation, useParams } from 'react-router';

function PathBreadcrumbs({ path, artifact }: { path: string; artifact?: Artifact }) {
    const base = artifact ? artifact.path : '';
    const crumbs = artifact
        ? [
              { label: `${artifact.groupId}:${artifact.artifactId}`, path: artifact.path },
              ...path
                  .slice(artifact.path.length)
                  .split('/')
                  .filter(Boolean)
                  .map((segment, i, rest) => ({ label: segment, path: join(base, ...rest.slice(0, i + 1)) })),
          ]
        : [];
    return (
        <Breadcrumb>
            <BreadcrumbList>
                <BreadcrumbItem>
                    <BreadcrumbLink render={<Link to="/browse" />}>Artifacts</BreadcrumbLink>
                </BreadcrumbItem>
                {crumbs.map((crumb, i) => {
                    const last = i === crumbs.length - 1;
                    return [
                        <BreadcrumbSeparator key={`sep-${i}`} />,
                        <BreadcrumbItem key={i}>
                            {last ? (
                                <BreadcrumbPage className="font-mono">{crumb.label}</BreadcrumbPage>
                            ) : (
                                <BreadcrumbLink className="font-mono" render={<Link to={browseUrl(crumb.path)} />}>
                                    {crumb.label}
                                </BreadcrumbLink>
                            )}
                        </BreadcrumbItem>,
                    ];
                })}
            </BreadcrumbList>
        </Breadcrumb>
    );
}

function NotFound({ path }: { path: string }) {
    const { data: me } = useMe();
    const location = useLocation();
    const group = path.split('/').join('.');

    return (
        <Empty className="border">
            <EmptyHeader>
                <EmptyMedia variant="icon">
                    <FolderSearch />
                </EmptyMedia>
                <EmptyTitle>Not found</EmptyTitle>
                <EmptyDescription>
                    {me
                        ? 'This path doesn’t exist or you don’t have access to it.'
                        : 'This path doesn’t exist, or it is private.'}
                </EmptyDescription>
            </EmptyHeader>
            <EmptyContent className="flex-row flex-wrap justify-center">
                <Button
                    variant="outline"
                    render={<Link to={`/browse?${new URLSearchParams({ q: group })}`} />}
                    nativeButton={false}
                >
                    <Search />
                    Search for <span className="font-mono">{group}</span>
                </Button>
                {!me && (
                    <Button render={<Link to="/login" state={{ from: location.pathname }} />} nativeButton={false}>
                        Log in to see private artifacts
                    </Button>
                )}
            </EmptyContent>
        </Empty>
    );
}

function ArtifactDirectory({ path }: { path: string }) {
    const { data: listing, error, isPending, isPlaceholderData } = useBrowse(path);
    const artifact = listing?.artifact;
    const rest = artifact ? path.slice(artifact.path.length).split('/').filter(Boolean) : [];
    usePageTitle(artifact && rest.length === 0 ? artifact.artifactId : path.split('/').at(-1));

    return (
        <>
            <PathBreadcrumbs path={path} artifact={artifact} />

            {isPending && (
                <div className="flex flex-col gap-2">
                    <Skeleton className="h-10" />
                    <Skeleton className="h-10" />
                    <Skeleton className="h-10" />
                </div>
            )}

            {isApiError(error, 404) && <NotFound path={path} />}
            {error && !isApiError(error, 404) && <p className="text-destructive">{error.message}</p>}

            {listing && !error && artifact && rest.length <= 1 && (
                <ArtifactDetail artifact={artifact} version={rest[0]} />
            )}
            {listing && !error && (!artifact || rest.length > 1) && (
                <div className={cn('transition-opacity', isPlaceholderData && 'opacity-60')}>
                    <EntryTable listing={listing} />
                </div>
            )}
        </>
    );
}

export function BrowsePage() {
    const path = (useParams()['*'] ?? '').split('/').filter(Boolean).join('/');

    if (path) return <ArtifactDirectory path={path} />;
    return (
        <>
            <PageHeader title="Browse" description="The artifacts you have access to, their versions and files." />
            <ArtifactList />
        </>
    );
}
