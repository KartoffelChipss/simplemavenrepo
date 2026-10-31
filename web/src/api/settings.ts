import { request } from '@/api/client.ts';
import { settingsKeys } from '@/api/keys.ts';
import type { Settings, UpdateSettingsRequest } from '@/api/types.ts';
import { queryOptions, useMutation, useQuery, useQueryClient } from '@tanstack/react-query';

export const DEFAULT_SERVER_NAME = 'simplemavenrepo';

export const getSettings = (signal?: AbortSignal) => request<Settings>('GET', '/settings', undefined, signal);
export const updateSettings = (body: UpdateSettingsRequest) => request<Settings>('PATCH', '/settings', body);

export const settingsQuery = () =>
    queryOptions({
        queryKey: settingsKeys.all,
        queryFn: ({ signal }) => getSettings(signal),
        staleTime: 5 * 60 * 1000,
    });

export function useSettings() {
    return useQuery(settingsQuery());
}

export function useServerName() {
    const { data } = useSettings();
    return data?.name ?? DEFAULT_SERVER_NAME;
}

export function useUpdateSettings() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: updateSettings,
        onSuccess: (settings) => queryClient.setQueryData(settingsKeys.all, settings),
    });
}
