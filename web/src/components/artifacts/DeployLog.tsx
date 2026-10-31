import { useDeploys } from '@/api/artifacts.ts';
import type { Artifact } from '@/api/types.ts';
import { UserAvatar } from '@/components/UserAvatar.tsx';
import { Badge } from '@/components/ui/badge';
import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty';
import {
    Pagination,
    PaginationContent,
    PaginationEllipsis,
    PaginationItem,
    PaginationLink,
    PaginationNext,
    PaginationPrevious,
} from '@/components/ui/pagination';
import { Skeleton } from '@/components/ui/skeleton';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { formatBytes, formatDateTime } from '@/lib/format.ts';
import { cn } from '@/lib/utils.ts';
import { History } from 'lucide-react';
import type { MouseEvent } from 'react';
import { useSearchParams } from 'react-router';

const PAGE_PARAM = 'deploys';

function visiblePages(current: number, count: number): (number | null)[] {
    const pages: (number | null)[] = [];
    for (let p = 1; p <= count; p++) {
        if (p === 1 || p === count || Math.abs(p - current) <= 1) {
            pages.push(p);
        } else if (pages[pages.length - 1] !== null) {
            pages.push(null);
        }
    }
    return pages;
}

function DeployPagination({ page, pageCount }: { page: number; pageCount: number }) {
    const [searchParams, setSearchParams] = useSearchParams();

    const paramsFor = (p: number) => {
        const params = new URLSearchParams(searchParams);
        if (p === 1) params.delete(PAGE_PARAM);
        else params.set(PAGE_PARAM, String(p));
        return params;
    };
    const href = (p: number) => `?${paramsFor(p).toString()}`;
    const go = (p: number) => (e: MouseEvent) => {
        e.preventDefault();
        if (p >= 1 && p <= pageCount && p !== page) {
            setSearchParams(paramsFor(p), { preventScrollReset: true });
        }
    };

    const isFirst = page <= 1;
    const isLast = page >= pageCount;

    return (
        <Pagination>
            <PaginationContent>
                <PaginationItem>
                    <PaginationPrevious
                        href={href(Math.max(page - 1, 1))}
                        onClick={go(page - 1)}
                        aria-disabled={isFirst}
                        className={cn(isFirst && 'pointer-events-none opacity-50')}
                    />
                </PaginationItem>
                {visiblePages(page, pageCount).map((p, i) =>
                    p === null ? (
                        <PaginationItem key={`gap-${i}`}>
                            <PaginationEllipsis />
                        </PaginationItem>
                    ) : (
                        <PaginationItem key={p}>
                            <PaginationLink href={href(p)} onClick={go(p)} isActive={p === page}>
                                {p}
                            </PaginationLink>
                        </PaginationItem>
                    )
                )}
                <PaginationItem>
                    <PaginationNext
                        href={href(Math.min(page + 1, pageCount))}
                        onClick={go(page + 1)}
                        aria-disabled={isLast}
                        className={cn(isLast && 'pointer-events-none opacity-50')}
                    />
                </PaginationItem>
            </PaginationContent>
        </Pagination>
    );
}

export function DeployLog({ artifact }: { artifact: Artifact }) {
    const [searchParams] = useSearchParams();
    const requested = Number(searchParams.get(PAGE_PARAM) ?? 1);
    const page = Number.isInteger(requested) && requested >= 1 ? requested : 1;

    const { data, error, isPending, isPlaceholderData } = useDeploys(artifact, page);
    const deploys = data?.deploys;
    const pageCount = data ? Math.max(Math.ceil(data.total / data.pageSize), 1) : 1;

    return (
        <>
            {error && <p className="text-destructive">{error.message}</p>}

            {isPending && (
                <div className="flex flex-col gap-2">
                    <Skeleton className="h-10" />
                    <Skeleton className="h-10" />
                </div>
            )}

            {data?.total === 0 && (
                <Empty className="border">
                    <EmptyHeader>
                        <EmptyMedia variant="icon">
                            <History />
                        </EmptyMedia>
                        <EmptyTitle>Nothing deployed yet</EmptyTitle>
                        <EmptyDescription>
                            Uploads of .jar and .pom files show up here, including failed ones.
                        </EmptyDescription>
                    </EmptyHeader>
                </Empty>
            )}

            {data && data.total > 0 && deploys?.length === 0 && (
                <p className="text-muted-foreground">There is no page {page}.</p>
            )}

            {deploys && deploys.length > 0 && (
                <div className={cn('rounded-lg border transition-opacity', isPlaceholderData && 'opacity-60')}>
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead>Time</TableHead>
                                <TableHead>User</TableHead>
                                <TableHead>Version</TableHead>
                                <TableHead>File</TableHead>
                                <TableHead>Result</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {deploys.map((deploy) => (
                                <TableRow key={deploy.id}>
                                    <TableCell className="whitespace-nowrap">
                                        {formatDateTime(deploy.createdAt)}
                                    </TableCell>
                                    <TableCell className="font-medium">
                                        <div className="flex items-center gap-2">
                                            <UserAvatar username={deploy.username} size={20} />
                                            {deploy.username}
                                        </div>
                                    </TableCell>
                                    <TableCell className="font-mono">{deploy.version}</TableCell>
                                    <TableCell className="font-mono text-muted-foreground">{deploy.file}</TableCell>
                                    <TableCell className="whitespace-normal">
                                        {deploy.success ? (
                                            <div className="flex items-center gap-2">
                                                <Badge variant="secondary">Deployed</Badge>
                                                {deploy.size !== null && (
                                                    <span className="text-muted-foreground">
                                                        {formatBytes(deploy.size)}
                                                    </span>
                                                )}
                                            </div>
                                        ) : (
                                            <div className="flex flex-col items-start gap-1">
                                                <Badge variant="destructive">Failed · {deploy.status}</Badge>
                                                {deploy.error && (
                                                    <span className="text-xs text-muted-foreground">
                                                        {deploy.error}
                                                    </span>
                                                )}
                                            </div>
                                        )}
                                    </TableCell>
                                </TableRow>
                            ))}
                        </TableBody>
                    </Table>
                </div>
            )}

            {pageCount > 1 && <DeployPagination page={page} pageCount={pageCount} />}
        </>
    );
}
