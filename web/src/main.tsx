import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import { RouterProvider } from 'react-router/dom';
import { Toaster } from '@/components/ui/toast';
import { queryClient } from '@/api/query-client.ts';
import { QueryClientProvider } from '@tanstack/react-query';
import { ReactQueryDevtools } from '@tanstack/react-query-devtools';
import { router } from './router.tsx';
import './index.css';
import './theme.ts';

createRoot(document.getElementById('root')!).render(
    <StrictMode>
        <QueryClientProvider client={queryClient}>
            <Toaster>
                <RouterProvider router={router} />
            </Toaster>
            <ReactQueryDevtools buttonPosition="bottom-right" />
        </QueryClientProvider>
    </StrictMode>
);
