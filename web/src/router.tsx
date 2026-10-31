import { createBrowserRouter } from 'react-router';
import { Layout } from './components/Layout.tsx';
import { RequireAdmin, RequireAuth } from './components/RequireAuth.tsx';
import { ErrorPage } from './pages/ErrorPage.tsx';

export const router = createBrowserRouter([
    { path: '/login', lazy: async () => ({ Component: (await import('./pages/LoginPage.tsx')).LoginPage }) },
    {
        element: <Layout />,
        errorElement: <ErrorPage />,
        HydrateFallback: () => null,
        children: [
            { index: true, lazy: async () => ({ Component: (await import('./pages/HomePage.tsx')).HomePage }) },
            {
                path: 'browse/*',
                lazy: async () => ({ Component: (await import('./pages/BrowsePage.tsx')).BrowsePage }),
            },
            {
                path: 'docs',
                lazy: async () => ({ Component: (await import('./pages/JavadocIndexPage.tsx')).JavadocIndexPage }),
            },
            {
                path: 'docs/*',
                lazy: async () => ({ Component: (await import('./pages/JavadocPage.tsx')).JavadocPage }),
            },
            {
                element: <RequireAuth />,
                children: [
                    {
                        path: 'tokens',
                        lazy: async () => ({ Component: (await import('./pages/TokensPage.tsx')).TokensPage }),
                    },
                    {
                        element: <RequireAdmin />,
                        children: [
                            {
                                path: 'artifacts',
                                lazy: async () => ({
                                    Component: (await import('./pages/ArtifactsPage.tsx')).ArtifactsPage,
                                }),
                            },
                            {
                                path: 'artifacts/:groupId/:artifactId',
                                lazy: async () => ({
                                    Component: (await import('./pages/ArtifactPage.tsx')).ArtifactPage,
                                }),
                            },
                            {
                                path: 'users',
                                lazy: async () => ({ Component: (await import('./pages/UsersPage.tsx')).UsersPage }),
                            },
                            {
                                path: 'users/:username',
                                lazy: async () => ({ Component: (await import('./pages/UserPage.tsx')).UserPage }),
                            },
                            {
                                path: 'teams',
                                lazy: async () => ({ Component: (await import('./pages/TeamsPage.tsx')).TeamsPage }),
                            },
                            {
                                path: 'teams/:name',
                                lazy: async () => ({ Component: (await import('./pages/TeamPage.tsx')).TeamPage }),
                            },
                            {
                                path: 'settings',
                                lazy: async () => ({
                                    Component: (await import('./pages/SettingsPage.tsx')).SettingsPage,
                                }),
                            },
                        ],
                    },
                ],
            },
            {
                path: '*',
                lazy: async () => ({ Component: (await import('./pages/NotFoundPage.tsx')).NotFoundPage }),
            },
        ],
    },
]);
