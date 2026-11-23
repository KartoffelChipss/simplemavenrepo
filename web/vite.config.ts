import tailwindcss from '@tailwindcss/vite';
import path from 'path';
import react from '@vitejs/plugin-react';
import { defineConfig } from 'vite';

const backend = process.env.SMR_BACKEND ?? 'http://localhost:4321';

let repositories: { names: Set<string>; fetchedAt: number } | undefined;
async function repositoryNames(): Promise<Set<string>> {
    if (repositories && Date.now() - repositories.fetchedAt < 5000) return repositories.names;
    let names = new Set<string>();
    try {
        const res = await fetch(`${backend}/api/repositories`);
        if (res.ok) names = new Set(((await res.json()) as { name: string }[]).map((r) => r.name));
    } catch {
        // Backend not running
    }
    repositories = { names, fetchedAt: Date.now() };
    return names;
}

// https://vite.dev/config/
export default defineConfig({
    plugins: [react(), tailwindcss()],
    resolve: {
        alias: {
            '@': path.resolve(import.meta.dirname, './src'),
        },
    },
    build: {
        outDir: '../server/internal/webui/dist',
        emptyOutDir: true,
    },
    server: {
        proxy: {
            '/api': backend,
            '/javadoc': backend,
            '^/[^/]+/': {
                target: backend,
                bypass: async (req) => {
                    const name = decodeURIComponent((req.url ?? '').split('/')[1] ?? '');
                    return (await repositoryNames()).has(name) ? undefined : req.url;
                },
            },
        },
    },
});
