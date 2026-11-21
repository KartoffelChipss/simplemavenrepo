import { request, seg } from '@/api/client.ts';
import { repoKeys } from '@/api/keys.ts';
import type { CreateRepositoryRequest, Repository, UpdateRepositoryRequest } from '@/api/types.ts';
import { queryOptions, useMutation, useQuery, useQueryClient } from '@tanstack/react-query';

export const listRepositories = (signal?: AbortSignal) =>
    request<Repository[]>('GET', '/repositories', undefined, signal);
export const getRepository = (name: string, signal?: AbortSignal) =>
    request<Repository>('GET', `/repositories/${seg(name)}`, undefined, signal);
export const createRepository = (body: CreateRepositoryRequest) => request<Repository>('POST', '/repositories', body);
export const updateRepository = (name: string, body: UpdateRepositoryRequest) =>
    request<Repository>('PATCH', `/repositories/${seg(name)}`, body);
export const deleteRepository = (name: string, deleteFiles: boolean) =>
    request<void>('DELETE', `/repositories/${seg(name)}${deleteFiles ? '?deleteFiles=true' : ''}`);

export const repositoriesQuery = () =>
    queryOptions({
        queryKey: repoKeys.list(),
        queryFn: ({ signal }) => listRepositories(signal),
        staleTime: 60 * 1000,
    });

export function useRepositories() {
    return useQuery(repositoriesQuery());
}

export function useCreateRepository() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: createRepository,
        onSuccess: () => queryClient.invalidateQueries({ queryKey: repoKeys.all }),
    });
}

export function useUpdateRepository() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: ({ name, ...body }: UpdateRepositoryRequest & { name: string }) => updateRepository(name, body),
        onSuccess: () => queryClient.invalidateQueries({ queryKey: repoKeys.all }),
    });
}

export function useDeleteRepository() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: ({ name, deleteFiles }: { name: string; deleteFiles: boolean }) =>
            deleteRepository(name, deleteFiles),
        onSuccess: () => queryClient.invalidateQueries({ queryKey: repoKeys.all }),
    });
}

export const versionPolicyLabels = {
    release: 'Releases only',
    snapshot: 'Snapshots only',
    mixed: 'Releases and snapshots',
} as const;
