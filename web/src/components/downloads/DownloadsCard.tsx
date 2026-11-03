import { useArtifactDownloads } from '@/api/artifacts.ts';
import type { Artifact, DownloadDays } from '@/api/types.ts';
import { DownloadsChart, DownloadTotals, PeriodSelect } from '@/components/downloads/DownloadsChart.tsx';
import { Card, CardAction, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { useState } from 'react';

export const DOWNLOADS_DESCRIPTION =
    'Counts each fetch of a version’s main file, like the .jar. POMs, checksums, sources and javadoc don’t count.';

/** Downloads of one artifact over time and per version. */
export function DownloadsCard({ artifact }: { artifact: Artifact }) {
    const [days, setDays] = useState<DownloadDays>(30);
    const { data, error, isPending } = useArtifactDownloads(artifact, days);

    return (
        <Card>
            <CardHeader>
                <CardTitle>Downloads</CardTitle>
                <CardDescription>{DOWNLOADS_DESCRIPTION}</CardDescription>
                <CardAction>
                    <PeriodSelect days={days} onChange={setDays} />
                </CardAction>
            </CardHeader>
            <CardContent className="flex flex-col gap-6">
                {isPending && <Skeleton className="h-72" />}
                {error && <p className="text-destructive">{error.message}</p>}
                {data && (
                    <>
                        <DownloadTotals periodTotal={data.periodTotal} total={data.total} days={data.days} />
                        <DownloadsChart daily={data.daily} />
                        {data.versions.length > 0 && (
                            <div className="rounded-lg border">
                                <Table>
                                    <TableHeader>
                                        <TableRow>
                                            <TableHead>Version</TableHead>
                                            <TableHead className="text-right">Downloads</TableHead>
                                        </TableRow>
                                    </TableHeader>
                                    <TableBody>
                                        {data.versions.map((v) => (
                                            <TableRow key={v.version}>
                                                <TableCell className="font-mono text-sm">{v.version}</TableCell>
                                                <TableCell className="text-right tabular-nums">
                                                    {v.count.toLocaleString()}
                                                </TableCell>
                                            </TableRow>
                                        ))}
                                    </TableBody>
                                </Table>
                            </div>
                        )}
                    </>
                )}
            </CardContent>
        </Card>
    );
}
