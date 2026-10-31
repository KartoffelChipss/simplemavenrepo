import { PageHeader } from '@/components/PageHeader.tsx';
import { useDocumentTitle } from '@/title.ts';
import { isRouteErrorResponse, useRouteError } from 'react-router';

export function ErrorPage() {
    const error = useRouteError();
    useDocumentTitle('Error');
    const message = isRouteErrorResponse(error)
        ? `${error.status} ${error.statusText}`
        : error instanceof Error
          ? error.message
          : String(error);

    return (
        <main className="p-6">
            <PageHeader
                title="Something went wrong"
                description={<span className="text-destructive">{message}</span>}
            />
        </main>
    );
}
