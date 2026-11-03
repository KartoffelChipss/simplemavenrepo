import { request } from '@/api/client.ts';
import { statsKeys } from '@/api/keys.ts';
import type { DownloadDays, DownloadStats, StorageStats } from '@/api/types.ts';
import { keepPreviousData, queryOptions, useQuery } from '@tanstack/react-query';

export const getStorageStats = (signal?: AbortSignal) =>
    request<StorageStats>('GET', '/stats/storage', undefined, signal);

export const storageStatsQuery = () =>
    queryOptions({
        queryKey: statsKeys.storage(),
        queryFn: ({ signal }) => getStorageStats(signal),
    });

export function useStorageStats({ enabled = true }: { enabled?: boolean } = {}) {
    return useQuery({ ...storageStatsQuery(), enabled });
}

export const getDownloadStats = (days: DownloadDays, signal?: AbortSignal) =>
    request<DownloadStats>('GET', `/stats/downloads?days=${days}`, undefined, signal);

export const downloadStatsQuery = (days: DownloadDays) =>
    queryOptions({
        queryKey: statsKeys.downloads(days),
        queryFn: ({ signal }) => getDownloadStats(days, signal),
        placeholderData: keepPreviousData,
    });

export function useDownloadStats(days: DownloadDays) {
    return useQuery(downloadStatsQuery(days));
}
