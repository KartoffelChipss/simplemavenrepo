import { request } from '@/api/client.ts';
import { grantKeys, meKeys, teamKeys, userKeys } from '@/api/keys.ts';
import type { CreateGrantRequest, Grant, GrantFilter } from '@/api/types.ts';
import { queryOptions, useMutation, useQuery, useQueryClient, type QueryClient } from '@tanstack/react-query';

export function listGrants(filter: GrantFilter = {}, signal?: AbortSignal) {
    const params = new URLSearchParams();
    if (filter.user) params.set('user', filter.user);
    if (filter.team) params.set('team', filter.team);
    const query = params.size ? `?${params}` : '';
    return request<Grant[]>('GET', `/grants${query}`, undefined, signal);
}

export const createGrant = (body: CreateGrantRequest) => request<Grant>('POST', '/grants', body);
export const deleteGrant = (id: number) => request<void>('DELETE', `/grants/${id}`);

export const grantsQuery = (filter: GrantFilter = {}) =>
    queryOptions({
        queryKey: grantKeys.list(filter),
        queryFn: ({ signal }) => listGrants(filter, signal),
    });

export function useGrants(filter: GrantFilter = {}) {
    return useQuery(grantsQuery(filter));
}

function invalidateGrants(queryClient: QueryClient) {
    return Promise.all([
        queryClient.invalidateQueries({ queryKey: grantKeys.all }),
        queryClient.invalidateQueries({ queryKey: userKeys.all }),
        queryClient.invalidateQueries({ queryKey: teamKeys.all }),
        queryClient.invalidateQueries({ queryKey: meKeys.all, exact: true }),
    ]);
}

export function useCreateGrant() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: createGrant,
        onSuccess: () => invalidateGrants(queryClient),
    });
}

export function useDeleteGrant() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: deleteGrant,
        onSuccess: () => invalidateGrants(queryClient),
    });
}
