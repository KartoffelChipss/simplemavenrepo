import { isApiError, request } from '@/api/client.ts';
import { meKeys } from '@/api/keys.ts';
import type { ChangePasswordRequest, LoginRequest, UserDetails } from '@/api/types.ts';
import { queryOptions, useMutation, useQuery, useQueryClient } from '@tanstack/react-query';

export async function fetchMe(signal?: AbortSignal): Promise<UserDetails | null> {
    try {
        return await request<UserDetails>('GET', '/me', undefined, signal);
    } catch (err) {
        if (isApiError(err, 401)) return null;
        throw err;
    }
}

export const login = (body: LoginRequest) => request<UserDetails>('POST', '/auth/login', body);
export const logout = () => request<void>('POST', '/auth/logout');
export const changeMyPassword = (body: ChangePasswordRequest) => request<void>('PUT', '/me/password', body);

export const meQuery = () =>
    queryOptions({
        queryKey: meKeys.all,
        queryFn: ({ signal }) => fetchMe(signal),
    });

export function useMe() {
    return useQuery(meQuery());
}

export function useLogin() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: login,
        onSuccess: (me) => {
            queryClient.removeQueries({ predicate: (q) => q.queryKey[0] !== meKeys.all[0] });
            queryClient.setQueryData(meKeys.all, me);
        },
    });
}

export function useLogout() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: logout,
        onSuccess: () => {
            queryClient.removeQueries({ predicate: (q) => q.queryKey[0] !== meKeys.all[0] });
            queryClient.setQueryData(meKeys.all, null);
        },
    });
}

export function useChangeMyPassword() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: changeMyPassword,
        onSuccess: () => queryClient.invalidateQueries({ queryKey: meKeys.all, exact: true }),
    });
}
