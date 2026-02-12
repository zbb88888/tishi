import { defineConfig } from 'astro/config';
import tailwind from '@astrojs/tailwind';
import node from '@astrojs/node';

// https://astro.build/config
export default defineConfig({
    integrations: [tailwind()],
    output: 'server',
    adapter: node({ mode: 'standalone' }),
    site: 'https://tishi.dev',
    vite: {
        server: {
            proxy: {
                '/api': {
                    target: 'http://localhost:8080',
                    changeOrigin: true,
                },
            },
        },
    },
});
