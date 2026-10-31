import { isApiError } from '@/api/client.ts';
import { javadocPageUrl, useJavadoc } from '@/api/javadoc.ts';
import type { Javadoc } from '@/api/types.ts';
import { JavadocNav } from '@/components/javadoc/JavadocNav.tsx';
import { Button } from '@/components/ui/button';
import { Empty, EmptyContent, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Skeleton } from '@/components/ui/skeleton';
import { useTheme } from '@/theme.ts';
import { usePageTitle } from '@/title.ts';
import { ArrowLeft, BookX } from 'lucide-react';
import { useCallback, useEffect, useRef, useState } from 'react';
import { Link, useLocation, useNavigate, useParams, useSearchParams } from 'react-router';

const browseUrl = (path: string) => `/browse/${path.split('/').filter(Boolean).map(encodeURIComponent).join('/')}`;

interface FrameMessage {
    type: 'smr-javadoc';
    path: string;
    hash: string;
    title: string;
}

const isFrameMessage = (data: unknown): data is FrameMessage =>
    typeof data === 'object' && data !== null && (data as { type?: unknown }).type === 'smr-javadoc';

function isIndexed(javadoc: Javadoc, page: string) {
    const base = page.split('#')[0];
    return (
        javadoc.types.some((t) => t.url === base) ||
        javadoc.packages.some((p) => p.url === base) ||
        javadoc.pages.some((p) => p.url === base)
    );
}

function JavadocFrame({
    javadoc,
    versionPath,
    page,
    onNavigate,
}: {
    javadoc: Javadoc;
    versionPath: string;
    page: string;
    onNavigate: (page: string, title: string) => void;
}) {
    const frameRef = useRef<HTMLIFrameElement>(null);
    const shown = useRef<string | null>(null);
    const theme = useTheme();

    useEffect(() => {
        const frame = frameRef.current;
        const target = `${versionPath}|${page}`;
        if (!frame || shown.current === target) return;
        shown.current = target;
        frame.src = javadoc.baseUrl + page;
    }, [javadoc.baseUrl, versionPath, page]);

    useEffect(() => {
        frameRef.current?.contentWindow?.postMessage({ type: 'smr-javadoc', theme }, '*');
    }, [theme]);

    useEffect(() => {
        const onMessage = (e: MessageEvent) => {
            const frame = frameRef.current;
            if (!frame || e.source !== frame.contentWindow || !isFrameMessage(e.data)) return;
            frame.contentWindow?.postMessage({ type: 'smr-javadoc', theme }, '*');
            const reported = e.data.path.replace(/^\/javadoc\/[^/]+\//, '') + e.data.hash;
            shown.current = `${versionPath}|${reported}`;
            onNavigate(reported, e.data.title);
        };
        window.addEventListener('message', onMessage);
        return () => window.removeEventListener('message', onMessage);
    }, [theme, versionPath, onNavigate]);

    return (
        <iframe
            ref={frameRef}
            title="Javadoc"
            sandbox="allow-scripts allow-popups allow-popups-to-escape-sandbox"
            referrerPolicy="no-referrer"
            className="min-h-0 flex-1 bg-background"
        />
    );
}

function NotFound({ artifactPath }: { artifactPath: string }) {
    return (
        <Empty className="border">
            <EmptyHeader>
                <EmptyMedia variant="icon">
                    <BookX />
                </EmptyMedia>
                <EmptyTitle>No javadoc</EmptyTitle>
                <EmptyDescription>
                    This version has no javadoc jar, or it doesn’t exist or you don’t have access to it.
                </EmptyDescription>
            </EmptyHeader>
            <EmptyContent>
                <Button variant="outline" render={<Link to={browseUrl(artifactPath)} />}>
                    <ArrowLeft />
                    Back to the artifact
                </Button>
            </EmptyContent>
        </Empty>
    );
}

export function JavadocPage() {
    const versionPath = (useParams()['*'] ?? '').split('/').filter(Boolean).join('/');
    const artifactPath = versionPath.split('/').slice(0, -1).join('/');
    const { data: javadoc, error, isPending } = useJavadoc(versionPath);
    const [params, setParams] = useSearchParams();
    const location = useLocation();
    const navigate = useNavigate();

    const carried = (location.state as { page?: string } | null)?.page;
    const page =
        params.get('page') ??
        (javadoc && carried && isIndexed(javadoc, carried) ? carried : undefined) ??
        javadoc?.startPage ??
        '';

    const [frameTitle, setFrameTitle] = useState<string>();
    usePageTitle(frameTitle || (javadoc ? `${javadoc.artifact.artifactId} ${javadoc.version}` : undefined));

    const open = (url: string) => setParams({ page: url });
    const onNavigate = useCallback(
        (reported: string, title: string) => {
            setFrameTitle(title);
            setParams((p) => (p.get('page') === reported ? p : { page: reported }), { replace: true });
        },
        [setParams]
    );

    if (isPending) {
        return (
            <div className="flex h-[calc(100svh-6.5rem)] gap-4">
                <Skeleton className="hidden w-72 md:block" />
                <Skeleton className="flex-1" />
            </div>
        );
    }
    if (isApiError(error, 404)) return <NotFound artifactPath={artifactPath} />;
    if (error) return <p className="text-destructive">{error.message}</p>;

    const { artifact, version, versions } = javadoc;
    return (
        <div className="flex h-[calc(100svh-6.5rem)] min-h-96 flex-col gap-3">
            <div className="flex flex-wrap items-center gap-2">
                <Button
                    variant="ghost"
                    size="icon-sm"
                    nativeButton={false}
                    aria-label="Back to the artifact"
                    render={<Link to={browseUrl(versionPath)} />}
                >
                    <ArrowLeft />
                </Button>
                <h1 className="mr-auto truncate font-mono text-base font-semibold">
                    <span className="text-muted-foreground">{artifact.groupId}:</span>
                    {artifact.artifactId}
                </h1>
                <nav className="flex flex-wrap gap-1" aria-label="Javadoc pages">
                    {javadoc.pages.map((p) => (
                        <Button
                            key={p.url}
                            variant={page.split('#')[0] === p.url ? 'secondary' : 'ghost'}
                            size="sm"
                            onClick={() => open(p.url)}
                        >
                            {p.title}
                        </Button>
                    ))}
                </nav>
                <Select
                    value={version}
                    onValueChange={(v) => {
                        if (v && v !== version) {
                            void navigate(javadocPageUrl(`${artifactPath}/${v}`), { state: { page } });
                        }
                    }}
                >
                    <SelectTrigger size="sm" aria-label="Version" className="min-w-28 font-mono">
                        <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                        {versions.map((v) => (
                            <SelectItem key={v} value={v} className="font-mono">
                                {v}
                            </SelectItem>
                        ))}
                    </SelectContent>
                </Select>
            </div>
            <div className="flex min-h-0 flex-1 flex-col overflow-hidden rounded-xl border bg-card md:flex-row">
                <aside className="flex max-h-56 flex-col border-b md:max-h-none md:w-72 md:border-r md:border-b-0">
                    <JavadocNav javadoc={javadoc} page={page} onOpen={open} />
                </aside>
                <JavadocFrame javadoc={javadoc} versionPath={versionPath} page={page} onNavigate={onNavigate} />
            </div>
        </div>
    );
}
