import { Navigate, Outlet, useLocation } from 'react-router';
import { useAuth } from '../auth-context.ts';

export function RequireAuth() {
    const { user, loading } = useAuth();
    const location = useLocation();

    if (loading) return null;
    if (!user) return <Navigate to="/login" replace state={{ from: location.pathname }} />;
    return <Outlet />;
}

export function RequireAdmin() {
    const { user } = useAuth();

    if (!user?.isAdmin) return <p className="text-muted-foreground">You need admin access to view this page.</p>;
    return <Outlet />;
}
