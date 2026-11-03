import { javadocPageUrl } from '@/api/javadoc.ts';
import { Navigate, useParams, useSearchParams } from 'react-router';

export function JavadocPage() {
    const versionPath = useParams()['*'] ?? '';
    const page = useSearchParams()[0].get('page') ?? undefined;
    return <Navigate to={javadocPageUrl(versionPath, page)} replace />;
}
