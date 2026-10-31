import { isRouteErrorResponse, useRouteError } from 'react-router';

export function ErrorPage() {
    const error = useRouteError();
    const message = isRouteErrorResponse(error)
        ? `${error.status} ${error.statusText}`
        : error instanceof Error
          ? error.message
          : String(error);

    return (
        <main>
            <h1>Something went wrong</h1>
            <p className="error">{message}</p>
        </main>
    );
}
