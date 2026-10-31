import { Navigate, Outlet, useLocation } from 'react-router';
import { useMe } from '@/api/auth.ts';

export function RequireAuth() {
    const { data: user, isPending } = useMe();
    const location = useLocation();

    if (isPending) return null;
    if (!user) return <Navigate to="/login" replace state={{ from: location.pathname }} />;
    return <Outlet />;
}

export function RequireAdmin() {
    const { data: user } = useMe();

    if (!user?.isAdmin) return <p className="text-muted-foreground">You need admin access to view this page.</p>;
    return <Outlet />;
}
