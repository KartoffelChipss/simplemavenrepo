import { useDownloadStats } from '@/api/stats.ts';
import type { DownloadDays } from '@/api/types.ts';
import { artifactSettingsUrl, useBackState } from '@/components/artifacts/urls.ts';
import { DOWNLOADS_DESCRIPTION } from '@/components/downloads/DownloadsCard.tsx';
import { DownloadsChart, DownloadTotals, PeriodSelect } from '@/components/downloads/DownloadsChart.tsx';
import { Card, CardAction, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { useState } from 'react';
import { Link } from 'react-router';

/** Downloads of all artifacts over time and the most downloaded ones. */
export function DownloadStatsSection() {
    const backState = useBackState('Stats');
    const [days, setDays] = useState<DownloadDays>(30);
    const { data, error, isPending } = useDownloadStats(days);

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
                        <div className="flex flex-col gap-3">
                            <h3 className="font-medium">Most downloaded</h3>
                            {data.topArtifacts.length === 0 ? (
                                <p className="text-muted-foreground">No downloads in the last {data.days} days.</p>
                            ) : (
                                <ol className="flex flex-col divide-y rounded-lg border">
                                    {data.topArtifacts.map((artifact) => (
                                        <li
                                            key={artifact.id}
                                            className="flex items-baseline justify-between gap-4 px-3 py-2"
                                        >
                                            <Link
                                                to={artifactSettingsUrl(artifact)}
                                                state={backState}
                                                className="min-w-0 truncate font-mono text-sm hover:underline"
                                            >
                                                <span className="text-muted-foreground">{artifact.groupId}:</span>
                                                <span className="font-medium">{artifact.artifactId}</span>
                                            </Link>
                                            <span className="shrink-0 text-sm tabular-nums">
                                                {artifact.downloads.toLocaleString()}
                                            </span>
                                        </li>
                                    ))}
                                </ol>
                            )}
                        </div>
                    </>
                )}
            </CardContent>
        </Card>
    );
}
