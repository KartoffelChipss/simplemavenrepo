import { PageHeader } from '@/components/PageHeader.tsx';
import { Link } from 'react-router';

export function NotFoundPage() {
    return (
        <PageHeader
            title="Not found"
            description={
                <>
                    This page does not exist.{' '}
                    <Link to="/" className="text-foreground underline underline-offset-4">
                        Back to the repository
                    </Link>
                </>
            }
        />
    );
}
