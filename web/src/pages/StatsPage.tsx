import { PageHeader } from '@/components/PageHeader.tsx';
import { DownloadStatsSection } from '@/components/stats/DownloadStats.tsx';
import { StorageStatsSection } from '@/components/stats/StorageStats.tsx';

export function StatsPage() {
    return (
        <>
            <PageHeader title="Stats" description="An overview of how the repository is used." />
            <DownloadStatsSection />
            <StorageStatsSection />
        </>
    );
}
