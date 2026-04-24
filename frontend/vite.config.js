import { defineConfig } from 'vite';

export default defineConfig({
    root: '.',
    server: {
        port: 5173,
        open: true,
        host: true, // Listen on all local IPs
    },
    build: {
        outDir: 'dist',
        cssCodeSplit: true,
        rollupOptions: {
            // Treat chart.js as external — loaded via <script> tag
            external: ['/js/lib/chart.js'],
        },
    },
});
