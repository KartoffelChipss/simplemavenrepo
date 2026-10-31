import { Link, NavLink, Outlet, useNavigate } from 'react-router';
import { useAuth } from '../auth-context.ts';

export function Layout() {
    const { user, loading, logout } = useAuth();
    const navigate = useNavigate();

    async function onLogout() {
        await logout();
        await navigate('/');
    }

    return (
        <div className="app">
            <header className="topbar">
                <Link to="/" className="brand">
                    simplemavenrepo
                </Link>
                <nav>
                    <NavLink to="/" end>
                        Browse
                    </NavLink>
                    {user?.isAdmin && (
                        <>
                            <NavLink to="/artifacts">Artifacts</NavLink>
                            <NavLink to="/users">Users</NavLink>
                            <NavLink to="/teams">Teams</NavLink>
                        </>
                    )}
                    {user && <NavLink to="/tokens">Tokens</NavLink>}
                </nav>
                <div className="session">
                    {user ? (
                        <>
                            <span>{user.username}</span>
                            <button type="button" onClick={() => void onLogout()}>
                                Log out
                            </button>
                        </>
                    ) : (
                        !loading && <Link to="/login">Log in</Link>
                    )}
                </div>
            </header>
            <main>
                <Outlet />
            </main>
        </div>
    );
}
