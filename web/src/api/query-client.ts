import { isApiError } from '@/api/client.ts';
import { meKeys } from '@/api/keys.ts';
import { MutationCache, QueryCache, QueryClient } from '@tanstack/react-query';

function onError(err: unknown) {
    if (isApiError(err, 401)) queryClient.setQueryData(meKeys.all, null);
}

export const queryClient = new QueryClient({
    queryCache: new QueryCache({ onError }),
    mutationCache: new MutationCache({ onError }),
    defaultOptions: {
        queries: {
            staleTime: 30_000,
            retry: (failureCount, err) => !(isApiError(err) && err.status < 500) && failureCount < 2,
        },
    },
});
