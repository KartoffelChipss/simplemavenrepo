import { Link, NavLink, Outlet } from 'react-router';
import { useAuth } from '../auth-context.ts';

export function Layout() {
    const { user, logout } = useAuth();

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
                            <button type="button" onClick={logout}>
                                Log out
                            </button>
                        </>
                    ) : (
                        <Link to="/login">Log in</Link>
                    )}
                </div>
            </header>
            <main>
                <Outlet />
            </main>
        </div>
    );
}
