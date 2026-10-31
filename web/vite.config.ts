import tailwindcss from '@tailwindcss/vite';
import path from 'path';
import react from '@vitejs/plugin-react';
import { defineConfig } from 'vite';

const backend = process.env.SMR_BACKEND ?? 'http://localhost:4321';

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
            '/releases': backend,
            '/javadoc': backend,
        },
    },
});
