import { request } from '@/api/client.ts';
import { artifactKeys, browseKeys, javadocKeys, statsKeys } from '@/api/keys.ts';
import type { BrowseArtifact, BrowseListing, Pom, VulnerabilityReport } from '@/api/types.ts';
import {
    keepPreviousData,
    queryOptions,
    skipToken,
    useMutation,
    useQuery,
    useQueryClient,
} from '@tanstack/react-query';

const encodePath = (path: string) => path.split('/').filter(Boolean).map(encodeURIComponent).join('/');

export const listBrowseArtifacts = (signal?: AbortSignal) =>
    request<BrowseArtifact[]>('GET', '/browse', undefined, signal);

export function useBrowseArtifacts() {
    return useQuery({ queryKey: browseKeys.list(), queryFn: ({ signal }) => listBrowseArtifacts(signal) });
}

export const browse = (path: string, signal?: AbortSignal) =>
    request<BrowseListing>('GET', `/browse/${encodePath(path)}`, undefined, signal);

/** URL of a file, given as its path starting with the repository. */
export const repositoryFileUrl = (path: string) => `/${encodePath(path)}`;

export const browseQuery = (path: string) =>
    queryOptions({
        queryKey: browseKeys.path(path),
        queryFn: ({ signal }) => browse(path, signal),
    });

export const getPom = (versionPath: string, signal?: AbortSignal) =>
    request<Pom>('GET', `/pom/${encodePath(versionPath)}`, undefined, signal);

export function usePom(versionPath: string | undefined) {
    return useQuery({
        queryKey: browseKeys.pom(versionPath ?? ''),
        queryFn: versionPath ? ({ signal }) => getPom(versionPath, signal) : skipToken,
    });
}

export const getVulnerabilities = (versionPath: string, signal?: AbortSignal) =>
    request<VulnerabilityReport>('GET', `/vulnerabilities/${encodePath(versionPath)}`, undefined, signal);

export function useVulnerabilities(versionPath: string | undefined, enabled: boolean) {
    return useQuery({
        queryKey: browseKeys.vulnerabilities(versionPath ?? ''),
        queryFn: versionPath && enabled ? ({ signal }) => getVulnerabilities(versionPath, signal) : skipToken,
        staleTime: 10 * 60 * 1000, // server caches for a day, so we can keep it for a while
        retry: false,
    });
}

export function useBrowse(path: string) {
    return useQuery({ ...browseQuery(path), placeholderData: keepPreviousData });
}

export const badgeUrl = (groupId: string, artifactId: string) =>
    `/api/badge/${encodeURIComponent(groupId)}/${encodeURIComponent(artifactId)}`;

export type BadgePeriod = 'week' | 'month' | 'year' | 'all';

export const downloadsBadgeUrl = (groupId: string, artifactId: string, period: BadgePeriod) =>
    `${badgeUrl(groupId, artifactId)}/downloads${period === 'all' ? '' : `?period=${period}`}`;

export const deleteVersion = (versionPath: string) => request<void>('DELETE', `/versions/${encodePath(versionPath)}`);

export function useDeleteVersion() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: deleteVersion,
        onSuccess: (_, versionPath) => {
            queryClient.removeQueries({ queryKey: browseKeys.path(versionPath) });
            queryClient.removeQueries({ queryKey: browseKeys.pom(versionPath) });
            queryClient.removeQueries({ queryKey: browseKeys.vulnerabilities(versionPath) });
            void queryClient.invalidateQueries({ queryKey: browseKeys.all });
            void queryClient.invalidateQueries({ queryKey: javadocKeys.all });
            void queryClient.invalidateQueries({ queryKey: artifactKeys.all });
            void queryClient.invalidateQueries({ queryKey: statsKeys.all });
        },
    });
}
