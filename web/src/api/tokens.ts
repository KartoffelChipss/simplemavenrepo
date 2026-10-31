import { request, seg } from '@/api/client.ts';
import { meKeys, userKeys } from '@/api/keys.ts';
import type { CreatedToken, CreateTokenRequest, Token } from '@/api/types.ts';
import { queryOptions, useMutation, useQuery, useQueryClient } from '@tanstack/react-query';

// Own tokens

export const listMyTokens = (signal?: AbortSignal) => request<Token[]>('GET', '/me/tokens', undefined, signal);
export const createMyToken = (body: CreateTokenRequest) => request<CreatedToken>('POST', '/me/tokens', body);
export const deleteMyToken = (id: number) => request<void>('DELETE', `/me/tokens/${id}`);

export const myTokensQuery = () =>
    queryOptions({
        queryKey: meKeys.tokens(),
        queryFn: ({ signal }) => listMyTokens(signal),
    });

export function useMyTokens() {
    return useQuery(myTokensQuery());
}

export function useCreateMyToken() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: createMyToken,
        onSuccess: () => queryClient.invalidateQueries({ queryKey: meKeys.tokens() }),
    });
}

export function useDeleteMyToken() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: deleteMyToken,
        onSuccess: () => queryClient.invalidateQueries({ queryKey: meKeys.tokens() }),
    });
}

// Tokens of other users, admin only

const userTokensPath = (username: string) => `/users/${seg(username)}/tokens`;

export const listUserTokens = (username: string, signal?: AbortSignal) =>
    request<Token[]>('GET', userTokensPath(username), undefined, signal);
export const createUserToken = (username: string, body: CreateTokenRequest) =>
    request<CreatedToken>('POST', userTokensPath(username), body);
export const deleteUserToken = (username: string, id: number) =>
    request<void>('DELETE', `${userTokensPath(username)}/${id}`);

export const userTokensQuery = (username: string) =>
    queryOptions({
        queryKey: userKeys.tokens(username),
        queryFn: ({ signal }) => listUserTokens(username, signal),
    });

export function useUserTokens(username: string) {
    return useQuery(userTokensQuery(username));
}

export function useCreateUserToken(username: string) {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (body: CreateTokenRequest) => createUserToken(username, body),
        onSuccess: () =>
            Promise.all([
                queryClient.invalidateQueries({ queryKey: userKeys.tokens(username) }),
                queryClient.invalidateQueries({ queryKey: meKeys.tokens() }),
            ]),
    });
}

export function useDeleteUserToken(username: string) {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (id: number) => deleteUserToken(username, id),
        onSuccess: () =>
            Promise.all([
                queryClient.invalidateQueries({ queryKey: userKeys.tokens(username) }),
                queryClient.invalidateQueries({ queryKey: meKeys.tokens() }),
            ]),
    });
}
