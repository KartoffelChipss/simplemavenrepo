import { PageHeader } from '@/components/PageHeader.tsx';
import { usePageTitle } from '@/title.ts';
import { Link } from 'react-router';

export function NotFoundPage() {
    usePageTitle('Not found');

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
