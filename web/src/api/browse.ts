import { request } from '@/api/client.ts';
import { browseKeys } from '@/api/keys.ts';
import type { BrowseListing } from '@/api/types.ts';
import { keepPreviousData, queryOptions, useQuery } from '@tanstack/react-query';

const encodePath = (path: string) => path.split('/').filter(Boolean).map(encodeURIComponent).join('/');

/** Lists a directory of the repository. Fails with a 404 for paths that don't exist or that the caller can't read. */
export const browse = (path: string, signal?: AbortSignal) =>
    request<BrowseListing>('GET', `/browse/${encodePath(path)}`, undefined, signal);

/** URL of a file in the repository, for downloading it. */
export const repositoryFileUrl = (path: string) => `/releases/${encodePath(path)}`;

export const browseQuery = (path: string) =>
    queryOptions({
        queryKey: browseKeys.path(path),
        queryFn: ({ signal }) => browse(path, signal),
    });

export function useBrowse(path: string) {
    // Keeps showing the previous directory while the next one loads, so navigating doesn't flash a skeleton.
    return useQuery({ ...browseQuery(path), placeholderData: keepPreviousData });
}

/** URL of the SVG badge showing the latest version of an artifact. */
export const badgeUrl = (groupId: string, artifactId: string) =>
    `/api/badge/${encodeURIComponent(groupId)}/${encodeURIComponent(artifactId)}`;

export type BadgePeriod = 'week' | 'month' | 'year' | 'all';

export const downloadsBadgeUrl = (groupId: string, artifactId: string, period: BadgePeriod) =>
    `${badgeUrl(groupId, artifactId)}/downloads${period === 'all' ? '' : `?period=${period}`}`;
