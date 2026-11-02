import { useStorageStats } from '@/api/stats.ts';
import type { StorageStats } from '@/api/types.ts';
import { artifactSettingsUrl } from '@/components/artifacts/urls.ts';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { formatBytes } from '@/lib/format.ts';
import { Link } from 'react-router';

function Stat({ label, value, hint }: { label: string; value: string; hint?: string }) {
    return (
        <div className="flex flex-col gap-1 rounded-lg border p-3">
            <span className="text-muted-foreground">{label}</span>
            <span className="font-heading text-2xl font-medium tabular-nums">{value}</span>
            {hint && <span className="text-xs text-muted-foreground">{hint}</span>}
        </div>
    );
}

function LargestArtifacts({ stats }: { stats: StorageStats }) {
    const artifacts = stats.largestArtifacts.filter((a) => a.size > 0);
    if (artifacts.length === 0) {
        return <p className="text-muted-foreground">No artifact has been published yet.</p>;
    }

    return (
        <ol className="flex flex-col gap-3">
            {artifacts.map((artifact) => {
                const share = stats.totalSize > 0 ? (artifact.size / stats.totalSize) * 100 : 0;
                return (
                    <li key={artifact.id} className="flex flex-col gap-1.5">
                        <div className="flex items-baseline justify-between gap-4">
                            <Link
                                to={artifactSettingsUrl(artifact)}
                                className="min-w-0 truncate font-mono text-sm hover:underline"
                            >
                                <span className="text-muted-foreground">{artifact.groupId}:</span>
                                <span className="font-medium">{artifact.artifactId}</span>
                            </Link>
                            <span className="shrink-0 text-sm tabular-nums">
                                {formatBytes(artifact.size)}
                                <span className="ml-2 inline-block w-12 text-right text-muted-foreground">
                                    {share < 1 ? '<1' : Math.round(share)}%
                                </span>
                            </span>
                        </div>
                        <div className="h-1.5 overflow-hidden rounded-full bg-muted">
                            <div
                                className="h-full rounded-full bg-primary"
                                // At least 1% so artifacts shown as "<1%" still get a visible sliver
                                style={{ width: `${Math.max(share, 1)}%` }}
                            />
                        </div>
                    </li>
                );
            })}
        </ol>
    );
}

export function StorageStatsSection() {
    const { data: stats, error, isPending } = useStorageStats();

    return (
        <Card>
            <CardHeader>
                <CardTitle>Storage</CardTitle>
                <CardDescription>Disk space used by the files in the repository.</CardDescription>
            </CardHeader>
            <CardContent className="flex flex-col gap-6">
                {isPending && (
                    <>
                        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
                            {Array.from({ length: 4 }, (_, i) => (
                                <Skeleton key={i} className="h-20" />
                            ))}
                        </div>
                        <Skeleton className="h-32" />
                    </>
                )}
                {error && <p className="text-destructive">{error.message}</p>}
                {stats && (
                    <>
                        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
                            <Stat label="Total" value={formatBytes(stats.totalSize)} />
                            <Stat label="Files" value={stats.fileCount.toLocaleString()} />
                            <Stat label="Artifacts" value={stats.artifactCount.toLocaleString()} />
                            <Stat
                                label="Unregistered"
                                value={formatBytes(stats.unregisteredSize)}
                                hint="Files of no registered artifact"
                            />
                        </div>
                        <div className="flex flex-col gap-3">
                            <h3 className="font-medium">Largest artifacts</h3>
                            <LargestArtifacts stats={stats} />
                        </div>
                    </>
                )}
            </CardContent>
        </Card>
    );
}
