import { createBrowserRouter } from 'react-router';
import { Layout } from './components/Layout.tsx';
import { RequireAdmin, RequireAuth } from './components/RequireAuth.tsx';
import { ArtifactsPage } from './pages/ArtifactsPage.tsx';
import { BrowsePage } from './pages/BrowsePage.tsx';
import { ErrorPage } from './pages/ErrorPage.tsx';
import { LoginPage } from './pages/LoginPage.tsx';
import { NotFoundPage } from './pages/NotFoundPage.tsx';
import { TeamsPage } from './pages/TeamsPage.tsx';
import { TokensPage } from './pages/TokensPage.tsx';
import { UserPage } from './pages/UserPage.tsx';
import { UsersPage } from './pages/UsersPage.tsx';

export const router = createBrowserRouter([
    { path: '/login', element: <LoginPage /> },
    {
        element: <Layout />,
        errorElement: <ErrorPage />,
        children: [
            { index: true, element: <BrowsePage /> },
            { path: 'browse/*', element: <BrowsePage /> },
            {
                element: <RequireAuth />,
                children: [
                    { path: 'tokens', element: <TokensPage /> },
                    {
                        element: <RequireAdmin />,
                        children: [
                            { path: 'artifacts', element: <ArtifactsPage /> },
                            { path: 'users', element: <UsersPage /> },
                            { path: 'users/:username', element: <UserPage /> },
                            { path: 'teams', element: <TeamsPage /> },
                        ],
                    },
                ],
            },
            { path: '*', element: <NotFoundPage /> },
        ],
    },
]);
