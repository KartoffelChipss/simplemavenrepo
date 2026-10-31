import { Link } from 'react-router';

export function NotFoundPage() {
    return (
        <>
            <h1>Not found</h1>
            <p>
                This page does not exist. <Link to="/">Back to the repository</Link>
            </p>
        </>
    );
}
