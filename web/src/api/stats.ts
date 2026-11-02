import { request } from '@/api/client.ts';
import { statsKeys } from '@/api/keys.ts';
import type { StorageStats } from '@/api/types.ts';
import { queryOptions, useQuery } from '@tanstack/react-query';

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
