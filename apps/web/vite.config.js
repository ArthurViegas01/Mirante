import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

const apiTarget = process.env.API_URL || 'http://localhost:8080';

// Em Docker (Windows/macOS) os eventos de arquivo não cruzam o bind mount, então
// o watcher do vite só percebe edições do host com polling. Ligado via
// CHOKIDAR_USEPOLLING (definido no docker-compose); fica desligado quando rodando
// direto no host, para não gastar CPU à toa.
const usePolling = process.env.CHOKIDAR_USEPOLLING === 'true';

export default defineConfig({
	plugins: [sveltekit()],
	server: {
		port: 5173,
		watch: usePolling ? { usePolling: true, interval: 100 } : undefined,
		proxy: {
			'/api': { target: apiTarget, changeOrigin: true },
			'/healthz': { target: apiTarget, changeOrigin: true }
		}
	}
});
