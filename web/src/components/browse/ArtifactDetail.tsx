import { useMe } from '@/api/auth.ts';
import { browseQuery, useBrowse, usePom } from '@/api/browse.ts';
import { isApiError } from '@/api/client.ts';
import type { Artifact, BrowseEntry, Pom, PomDependency } from '@/api/types.ts';
import { artifactSettingsUrl } from '@/components/artifacts/urls.ts';
import { EntryTable } from '@/components/browse/EntryTable.tsx';
import { browseUrl, isJavadocJar } from '@/components/browse/paths.ts';
import { UsageSnippets } from '@/components/browse/UsageSnippets.tsx';
import { JavadocViewer } from '@/components/javadoc/JavadocViewer.tsx';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Skeleton } from '@/components/ui/skeleton';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { formatDate } from '@/lib/format.ts';
import { cn } from '@/lib/utils';
import { keepPreviousData, useQuery } from '@tanstack/react-query';
import { Globe, Lock, PackageOpen, Settings } from 'lucide-react';
import type { ReactNode } from 'react';
import { Link, useNavigate, useSearchParams } from 'react-router';
import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '../ui/empty';

const TABS = ['overview', 'versions', 'dependencies', 'javadoc'] as const;
type Tab = (typeof TABS)[number];

const SCOPE_ORDER = ['compile', 'runtime', 'provided', 'system', 'test'];

const isHttpUrl = (url: string) => /^https?:\/\//i.test(url);

function ExternalLink({ href, children }: { href: string; children?: ReactNode }) {
    if (!isHttpUrl(href)) return <span className="break-all">{children ?? href}</span>;
    return (
        <a href={href} target="_blank" rel="noreferrer noopener" className="break-all hover:underline">
            {children ?? href.replace(/^https?:\/\//i, '').replace(/\/$/, '')}
        </a>
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

function Section({ title, children }: { title: string; children: ReactNode }) {
    return (
        <section className="flex flex-col gap-3">
            <h2 className="text-base font-semibold">{title}</h2>
            {children}
        </section>
    );
}

function DetailRow({ label, children }: { label: string; children: ReactNode }) {
    return (
        <div className="flex flex-col gap-0.5">
            <dt className="text-xs font-medium text-muted-foreground uppercase">{label}</dt>
            <dd className="text-sm">{children}</dd>
        </div>
    );
}

function Details({
    version,
    latest,
    published,
    pom,
    pomPending,
}: {
    version: string;
    latest?: string;
    published?: string;
    pom?: Pom;
    pomPending: boolean;
}) {
    return (
        <dl className="flex flex-col gap-4 rounded-lg border p-4">
            <DetailRow label="Version">
                <span className="flex items-center gap-2 font-mono">
                    {version}
                    {version === latest && <Badge variant="secondary">Latest</Badge>}
                </span>
            </DetailRow>
            {published && <DetailRow label="Published">{formatDate(published)}</DetailRow>}
            {pomPending && (
                <>
                    <Skeleton className="h-9" />
                    <Skeleton className="h-9" />
                </>
            )}
            {pom && (
                <>
                    <DetailRow label="Packaging">
                        <span className="font-mono">{pom.packaging}</span>
                    </DetailRow>
                    {pom.licenses.length > 0 && (
                        <DetailRow label={pom.licenses.length === 1 ? 'License' : 'Licenses'}>
                            <ul className="flex flex-col gap-0.5">
                                {pom.licenses.map((l) => (
                                    <li key={l.name}>
                                        {l.url ? <ExternalLink href={l.url}>{l.name}</ExternalLink> : l.name}
                                    </li>
                                ))}
                            </ul>
                        </DetailRow>
                    )}
                    {pom.url && (
                        <DetailRow label="Homepage">
                            <ExternalLink href={pom.url} />
                        </DetailRow>
                    )}
                    {pom.scm?.url && (
                        <DetailRow label="Source">
                            <ExternalLink href={pom.scm.url} />
                        </DetailRow>
                    )}
                    {pom.issueManagement && (
                        <DetailRow label="Issues">
                            <ExternalLink href={pom.issueManagement.url} />
                        </DetailRow>
                    )}
                    {pom.developers.length > 0 && (
                        <DetailRow label={pom.developers.length === 1 ? 'Developer' : 'Developers'}>
                            <ul className="flex flex-col gap-1">
                                {pom.developers.map((d, i) => {
                                    const name = d.name || d.id || d.email;
                                    return (
                                        <li key={i}>
                                            {d.url ? <ExternalLink href={d.url}>{name}</ExternalLink> : name}
                                            {d.organization && (
                                                <span className="text-muted-foreground"> · {d.organization}</span>
                                            )}
                                        </li>
                                    );
                                })}
                            </ul>
                        </DetailRow>
                    )}
                    {pom.parent && (
                        <DetailRow label="Parent">
                            <span className="font-mono break-all">
                                {pom.parent.groupId}:{pom.parent.artifactId}:{pom.parent.version}
                            </span>
                        </DetailRow>
                    )}
                    {pom.unresolved.length > 0 && (
                        <p className="text-xs text-muted-foreground">
                            Not in this repository, so details inherited from it are missing:{' '}
                            <span className="font-mono break-all">{pom.unresolved.join(', ')}</span>
                        </p>
                    )}
                </>
            )}
        </dl>
    );
}

function VersionsTable({
    artifact,
    versions,
    current,
    latest,
}: {
    artifact: Artifact;
    versions: BrowseEntry[];
    current?: string;
    latest?: string;
}) {
    const navigate = useNavigate();

    if (versions.length === 0) {
        return <p className="text-sm text-muted-foreground">No version has been published yet.</p>;
    }
    return (
        <div className="rounded-lg border">
            <Table>
                <TableHeader>
                    <TableRow>
                        <TableHead>Version</TableHead>
                        <TableHead>Published</TableHead>
                    </TableRow>
                </TableHeader>
                <TableBody>
                    {versions.map((v) => {
                        const to = browseUrl(`${artifact.path}/${v.name}`);
                        return (
                            <TableRow
                                key={v.name}
                                className={cn('cursor-pointer', v.name === current && 'bg-muted/50')}
                                onClick={() => void navigate(to)}
                            >
                                <TableCell>
                                    <span className="flex items-center gap-2">
                                        <Link
                                            to={to}
                                            className="font-mono hover:underline"
                                            onClick={(e) => e.stopPropagation()}
                                        >
                                            {v.name}
                                        </Link>
                                        {v.name === latest && <Badge variant="secondary">Latest</Badge>}
                                        {v.name.endsWith('-SNAPSHOT') && <Badge variant="outline">Snapshot</Badge>}
                                    </span>
                                </TableCell>
                                <TableCell className="text-muted-foreground">
                                    {v.modifiedAt ? formatDate(v.modifiedAt) : ''}
                                </TableCell>
                            </TableRow>
                        );
                    })}
                </TableBody>
            </Table>
        </div>
    );
}

function DependenciesTable({ dependencies }: { dependencies: PomDependency[] }) {
    if (dependencies.length === 0) {
        return (
            <Empty>
                <EmptyHeader>
                    <EmptyMedia variant="icon">
                        <PackageOpen />
                    </EmptyMedia>
                    <EmptyTitle>No dependencies</EmptyTitle>
                    <EmptyDescription>This version has no dependencies.</EmptyDescription>
                </EmptyHeader>
            </Empty>
        );
    }
    const rank = (scope: string) => {
        const i = SCOPE_ORDER.indexOf(scope);
        return i < 0 ? SCOPE_ORDER.length : i;
    };
    const sorted = dependencies.toSorted((a, b) => rank(a.scope) - rank(b.scope));

    return (
        <div className="rounded-lg border">
            <Table>
                <TableHeader>
                    <TableRow>
                        <TableHead>Dependency</TableHead>
                        <TableHead>Version</TableHead>
                        <TableHead>Scope</TableHead>
                    </TableRow>
                </TableHeader>
                <TableBody>
                    {sorted.map((d, i) => {
                        const coordinates = (
                            <>
                                <span className="text-muted-foreground">{d.groupId}:</span>
                                {d.artifactId}
                                {d.classifier && <span className="text-muted-foreground">:{d.classifier}</span>}
                            </>
                        );
                        return (
                            <TableRow key={i}>
                                <TableCell className="font-mono whitespace-normal">
                                    {d.path ? (
                                        <Link to={browseUrl(d.path)} className="hover:underline">
                                            {coordinates}
                                        </Link>
                                    ) : (
                                        coordinates
                                    )}
                                    {d.type && (
                                        <span className="ml-2 font-sans text-xs text-muted-foreground">{d.type}</span>
                                    )}
                                </TableCell>
                                <TableCell className="font-mono">
                                    {d.version ?? <span className="font-sans text-muted-foreground">Managed</span>}
                                </TableCell>
                                <TableCell>
                                    <span className="flex gap-1">
                                        <Badge variant={d.scope === 'test' ? 'outline' : 'secondary'}>{d.scope}</Badge>
                                        {d.optional && <Badge variant="outline">optional</Badge>}
                                    </span>
                                </TableCell>
                            </TableRow>
                        );
                    })}
                </TableBody>
            </Table>
        </div>
    );
}

export function ArtifactDetail({ artifact, version: selected }: { artifact: Artifact; version?: string }) {
    const { data: me } = useMe();
    const navigate = useNavigate();
    const [searchParams, setSearchParams] = useSearchParams();
    const tabParam = searchParams.get('tab');
    const tab: Tab = TABS.find((t) => t === tabParam) ?? 'overview';

    const { data: root } = useBrowse(artifact.path);
    const versions = root?.entries.filter((e) => e.type === 'directory' && !e.artifact) ?? [];
    const latest = root?.latestVersion;
    const version = selected ?? latest ?? versions[0]?.name;
    const versionPath = version ? `${artifact.path}/${version}` : undefined;

    const { data: files, isPlaceholderData } = useQuery({
        ...browseQuery(versionPath ?? ''),
        enabled: versionPath !== undefined,
        placeholderData: keepPreviousData,
    });
    const { data: pom, error: pomError, isPending: pomPending } = usePom(versionPath);

    const description = artifact.description || pom?.description;
    const hasJavadoc = files?.entries.some(isJavadocJar) ?? false;
    const published = versions.find((v) => v.name === version)?.modifiedAt;

    const setTab = (t: string) => {
        const next = new URLSearchParams(searchParams);
        if (t === 'overview') next.delete('tab');
        else next.set('tab', t);
        if (t !== 'javadoc') next.delete('page');
        setSearchParams(next, { replace: true });
    };

    return (
        <div className="flex flex-col gap-6">
            <div className="flex flex-wrap items-start justify-between gap-4">
                <div className="flex min-w-0 flex-col gap-1">
                    <div className="flex flex-wrap items-center gap-2">
                        <h1 className="text-2xl font-semibold tracking-tight">{pom?.name || artifact.artifactId}</h1>
                        <Visibility isPublic={artifact.public} />
                    </div>
                    <p className="font-mono text-sm break-all text-muted-foreground">
                        {artifact.groupId}:{artifact.artifactId}
                    </p>
                    {description && <p className="text-muted-foreground">{description}</p>}
                </div>
                <div className="flex flex-wrap items-center gap-2">
                    {version && versions.length > 0 && (
                        <Select
                            value={version}
                            onValueChange={(v) => {
                                if (v && v !== version) {
                                    const next = new URLSearchParams(searchParams);
                                    const page = next.get('page') ?? undefined;
                                    next.delete('page');
                                    void navigate(
                                        { pathname: browseUrl(`${artifact.path}/${v}`), search: next.toString() },
                                        { state: { page } }
                                    );
                                }
                            }}
                        >
                            <SelectTrigger size="sm" aria-label="Version" className="min-w-28 font-mono">
                                <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
                                {versions.map((v) => (
                                    <SelectItem key={v.name} value={v.name} className="font-mono">
                                        {v.name}
                                    </SelectItem>
                                ))}
                            </SelectContent>
                        </Select>
                    )}
                    {me?.isAdmin && (
                        <Button
                            variant="outline"
                            size="sm"
                            render={<Link to={artifactSettingsUrl(artifact)} />}
                            nativeButton={false}
                        >
                            <Settings />
                            Settings
                        </Button>
                    )}
                </div>
            </div>

            <Tabs value={tab} onValueChange={setTab} className="gap-4">
                <TabsList variant="line" className="w-full justify-start border-b">
                    <TabsTrigger value="overview" className="flex-none">
                        Overview
                    </TabsTrigger>
                    <TabsTrigger value="versions" className="flex-none">
                        Versions
                        <span className="text-xs text-muted-foreground tabular-nums">{versions.length}</span>
                    </TabsTrigger>
                    <TabsTrigger value="dependencies" className="flex-none">
                        Dependencies
                        {pom && (
                            <span className="text-xs text-muted-foreground tabular-nums">
                                {pom.dependencies.length}
                            </span>
                        )}
                    </TabsTrigger>
                    {(hasJavadoc || tab === 'javadoc') && (
                        <TabsTrigger value="javadoc" className="flex-none">
                            Javadoc
                        </TabsTrigger>
                    )}
                </TabsList>

                <TabsContent value="overview">
                    {version ? (
                        <div className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_18rem]">
                            <div className="flex min-w-0 flex-col gap-6">
                                <Section title="Installation">
                                    <UsageSnippets artifact={artifact} version={version} />
                                </Section>
                                {files && (
                                    <Section title="Files">
                                        <div className={cn('transition-opacity', isPlaceholderData && 'opacity-60')}>
                                            <EntryTable listing={files} />
                                        </div>
                                    </Section>
                                )}
                            </div>
                            <aside>
                                <Details
                                    version={version}
                                    latest={latest}
                                    published={published}
                                    pom={pom}
                                    pomPending={pomPending && !pomError}
                                />
                            </aside>
                        </div>
                    ) : (
                        <p className="text-muted-foreground">No version has been published yet.</p>
                    )}
                </TabsContent>

                <TabsContent value="versions">
                    <VersionsTable artifact={artifact} versions={versions} current={version} latest={latest} />
                </TabsContent>

                <TabsContent value="dependencies">
                    {!version && <p className="text-muted-foreground">No version has been published yet.</p>}
                    {version && pomPending && !pomError && <Skeleton className="h-24" />}
                    {isApiError(pomError, 404) && <p className="text-muted-foreground">This version has no POM.</p>}
                    {pomError && !isApiError(pomError, 404) && <p className="text-destructive">{pomError.message}</p>}
                    {pom && <DependenciesTable dependencies={pom.dependencies} />}
                </TabsContent>

                <TabsContent value="javadoc">
                    {versionPath ? (
                        <JavadocViewer versionPath={versionPath} />
                    ) : (
                        <p className="text-muted-foreground">No version has been published yet.</p>
                    )}
                </TabsContent>
            </Tabs>
        </div>
    );
}
