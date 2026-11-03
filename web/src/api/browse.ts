import { request } from '@/api/client.ts';
import { browseKeys } from '@/api/keys.ts';
import type { BrowseArtifact, BrowseListing } from '@/api/types.ts';
import { keepPreviousData, queryOptions, useQuery } from '@tanstack/react-query';

const encodePath = (path: string) => path.split('/').filter(Boolean).map(encodeURIComponent).join('/');

export const listBrowseArtifacts = (signal?: AbortSignal) =>
    request<BrowseArtifact[]>('GET', '/browse', undefined, signal);

export function useBrowseArtifacts() {
    return useQuery({ queryKey: browseKeys.list(), queryFn: ({ signal }) => listBrowseArtifacts(signal) });
}

export const browse = (path: string, signal?: AbortSignal) =>
    request<BrowseListing>('GET', `/browse/${encodePath(path)}`, undefined, signal);

export const repositoryFileUrl = (path: string) => `/releases/${encodePath(path)}`;

export const browseQuery = (path: string) =>
    queryOptions({
        queryKey: browseKeys.path(path),
        queryFn: ({ signal }) => browse(path, signal),
    });

export function useBrowse(path: string) {
    return useQuery({ ...browseQuery(path), placeholderData: keepPreviousData });
}

export const badgeUrl = (groupId: string, artifactId: string) =>
    `/api/badge/${encodeURIComponent(groupId)}/${encodeURIComponent(artifactId)}`;

export type BadgePeriod = 'week' | 'month' | 'year' | 'all';

export const downloadsBadgeUrl = (groupId: string, artifactId: string, period: BadgePeriod) =>
    `${badgeUrl(groupId, artifactId)}/downloads${period === 'all' ? '' : `?period=${period}`}`;
