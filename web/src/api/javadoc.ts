import { request } from '@/api/client.ts';
import { javadocKeys } from '@/api/keys.ts';
import type { Javadoc, JavadocArtifact } from '@/api/types.ts';
import { useQuery } from '@tanstack/react-query';

const encodePath = (path: string) => path.split('/').filter(Boolean).map(encodeURIComponent).join('/');

/** Lists the artifacts the caller can read that have javadoc. */
export const listJavadoc = (signal?: AbortSignal) => request<JavadocArtifact[]>('GET', '/javadoc', undefined, signal);

export function useJavadocList() {
    return useQuery({ queryKey: javadocKeys.list(), queryFn: ({ signal }) => listJavadoc(signal) });
}

export const getJavadoc = (path: string, signal?: AbortSignal) =>
    request<Javadoc>('GET', `/javadoc/${encodePath(path)}`, undefined, signal);

export function useJavadoc(path: string) {
    return useQuery({
        queryKey: javadocKeys.path(path),
        queryFn: ({ signal }) => getJavadoc(path, signal),
        staleTime: 60 * 60 * 1000,
        refetchOnWindowFocus: false,
    });
}

export function javadocPageUrl(versionPath: string, page?: string) {
    const url = `/docs/${encodePath(versionPath)}`;
    return page ? `${url}?page=${encodeURIComponent(page)}` : url;
}
