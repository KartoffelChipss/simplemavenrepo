import { request, seg } from '@/api/client.ts';
import { grantKeys, meKeys, teamKeys, userKeys } from '@/api/keys.ts';
import type { CreateUserRequest, UpdateUserRequest, User, UserDetails } from '@/api/types.ts';
import { queryOptions, useMutation, useQuery, useQueryClient } from '@tanstack/react-query';

const userPath = (username: string) => `/users/${seg(username)}`;

export const listUsers = (signal?: AbortSignal) => request<User[]>('GET', '/users', undefined, signal);
export const getUser = (username: string, signal?: AbortSignal) =>
    request<UserDetails>('GET', userPath(username), undefined, signal);
export const createUser = (body: CreateUserRequest) => request<User>('POST', '/users', body);
export const updateUser = (username: string, body: UpdateUserRequest) =>
    request<User>('PATCH', userPath(username), body);
export const deleteUser = (username: string) => request<void>('DELETE', userPath(username));

export const usersQuery = () =>
    queryOptions({
        queryKey: userKeys.list(),
        queryFn: ({ signal }) => listUsers(signal),
    });

export const userQuery = (username: string) =>
    queryOptions({
        queryKey: userKeys.detail(username),
        queryFn: ({ signal }) => getUser(username, signal),
    });

export function useUsers() {
    return useQuery(usersQuery());
}

export function useUser(username: string) {
    return useQuery(userQuery(username));
}

export function useCreateUser() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: createUser,
        onSuccess: () => queryClient.invalidateQueries({ queryKey: userKeys.list() }),
    });
}

export function useUpdateUser() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: ({ username, ...body }: UpdateUserRequest & { username: string }) => updateUser(username, body),
        // Also refreshes the own user, an admin may change their own admin flag or password.
        onSuccess: () =>
            Promise.all([
                queryClient.invalidateQueries({ queryKey: userKeys.all }),
                queryClient.invalidateQueries({ queryKey: meKeys.all, exact: true }),
            ]),
    });
}

export function useDeleteUser() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: deleteUser,
        onSuccess: (_, username) => {
            queryClient.removeQueries({ queryKey: userKeys.detail(username) });
            // Deleting a user also removes their team memberships and grants.
            return Promise.all([
                queryClient.invalidateQueries({ queryKey: userKeys.list() }),
                queryClient.invalidateQueries({ queryKey: teamKeys.all }),
                queryClient.invalidateQueries({ queryKey: grantKeys.all }),
            ]);
        },
    });
}
