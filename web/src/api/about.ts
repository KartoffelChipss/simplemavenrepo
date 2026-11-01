import { request } from '@/api/client.ts';
import { aboutKeys } from '@/api/keys.ts';
import type { About } from '@/api/types.ts';
import { queryOptions, useQuery } from '@tanstack/react-query';

export const getAbout = (signal?: AbortSignal) => request<About>('GET', '/about', undefined, signal);

export const aboutQuery = () =>
    queryOptions({
        queryKey: aboutKeys.all,
        queryFn: ({ signal }) => getAbout(signal),
        staleTime: Infinity,
    });

export function useAbout({ enabled = true }: { enabled?: boolean } = {}) {
    return useQuery({ ...aboutQuery(), enabled });
}
