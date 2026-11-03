import { isApiError } from '@/api/client.ts';
import { useJavadoc } from '@/api/javadoc.ts';
import type { Javadoc } from '@/api/types.ts';
import { JavadocNav } from '@/components/javadoc/JavadocNav.tsx';
import { Button } from '@/components/ui/button';
import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty';
import { Skeleton } from '@/components/ui/skeleton';
import { useTheme } from '@/theme.ts';
import { usePageTitle } from '@/title.ts';
import { BookX } from 'lucide-react';
import { useCallback, useEffect, useLayoutEffect, useRef, useState } from 'react';
import { useLocation, useSearchParams } from 'react-router';

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

export function NoJavadoc() {
    return (
        <Empty>
            <EmptyHeader>
                <EmptyMedia variant="icon">
                    <BookX />
                </EmptyMedia>
                <EmptyTitle>No javadoc</EmptyTitle>
                <EmptyDescription>
                    This version has no javadoc jar. Publish one with the classifier{' '}
                    <code className="font-mono">javadoc</code> to browse it here.
                </EmptyDescription>
            </EmptyHeader>
        </Empty>
    );
}

const BOTTOM_GAP = 24;
const MIN_HEIGHT = 384;

function useRemainingHeight() {
    const [element, setElement] = useState<HTMLElement | null>(null);
    const [height, setHeight] = useState<number>();

    useLayoutEffect(() => {
        if (!element) return;
        const update = () => {
            const top = element.getBoundingClientRect().top + window.scrollY;
            setHeight(Math.max(MIN_HEIGHT, window.innerHeight - top - BOTTOM_GAP));
        };
        update();
        const observer = new ResizeObserver(update);
        observer.observe(document.body);
        window.addEventListener('resize', update);
        return () => {
            observer.disconnect();
            window.removeEventListener('resize', update);
        };
    }, [element]);

    return { ref: setElement, style: { height } };
}

export function JavadocViewer({ versionPath }: { versionPath: string }) {
    const { data: javadoc, error, isPending } = useJavadoc(versionPath);
    const [params, setParams] = useSearchParams();
    const location = useLocation();

    const carried = (location.state as { page?: string } | null)?.page;
    const page =
        params.get('page') ??
        (javadoc && carried && isIndexed(javadoc, carried) ? carried : undefined) ??
        javadoc?.startPage ??
        '';

    const [frameTitle, setFrameTitle] = useState<string>();
    usePageTitle(frameTitle);
    const fill = useRemainingHeight();

    const setPage = useCallback(
        (url: string, replace = false) =>
            setParams(
                (p) => {
                    if (p.get('page') === url) return p;
                    const next = new URLSearchParams(p);
                    next.set('page', url);
                    return next;
                },
                { replace }
            ),
        [setParams]
    );
    const onNavigate = useCallback(
        (reported: string, title: string) => {
            setFrameTitle(title);
            setPage(reported, true);
        },
        [setPage]
    );

    if (isPending) return <Skeleton {...fill} />;
    if (isApiError(error, 404)) return <NoJavadoc />;
    if (error) return <p className="text-destructive">{error.message}</p>;

    return (
        <div {...fill} className="flex flex-col gap-3">
            <nav className="flex flex-wrap gap-1" aria-label="Javadoc pages">
                {javadoc.pages.map((p) => (
                    <Button
                        key={p.url}
                        variant={page.split('#')[0] === p.url ? 'secondary' : 'ghost'}
                        size="sm"
                        onClick={() => setPage(p.url)}
                    >
                        {p.title}
                    </Button>
                ))}
            </nav>
            <div className="flex min-h-0 flex-1 flex-col overflow-hidden rounded-xl border bg-card md:flex-row">
                <aside className="flex max-h-56 flex-col border-b md:max-h-none md:w-72 md:border-r md:border-b-0">
                    <JavadocNav javadoc={javadoc} page={page} onOpen={setPage} />
                </aside>
                <JavadocFrame javadoc={javadoc} versionPath={versionPath} page={page} onNavigate={onNavigate} />
            </div>
        </div>
    );
}
