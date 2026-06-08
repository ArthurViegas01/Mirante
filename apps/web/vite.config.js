import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

const apiTarget = process.env.API_URL || 'http://localhost:8080';

export default defineConfig({
	plugins: [sveltekit()],
	server: {
		port: 5173,
		proxy: {
			'/api': { target: apiTarget, changeOrigin: true },
			'/healthz': { target: apiTarget, changeOrigin: true }
		}
	}
});
