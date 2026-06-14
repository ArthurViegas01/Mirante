// Proxy all /api/* and /healthz requests to the Go API at runtime.
// API_URL is set as a Fly secret (or env var) and read here server-side.
// This replaces Vite's dev-only proxy for production (adapter-node).

const API_URL = process.env.API_URL || 'http://localhost:8080';

export async function handle({ event, resolve }) {
	const { pathname } = event.url;

	if (pathname.startsWith('/api') || pathname === '/healthz') {
		const upstream = API_URL + event.url.pathname + event.url.search;

		/** @type {RequestInit} */
		const init = {
			method: event.request.method,
			headers: event.request.headers,
		};

		if (event.request.method !== 'GET' && event.request.method !== 'HEAD') {
			init.body = await event.request.arrayBuffer();
		}

		const resp = await fetch(upstream, init);

		return new Response(resp.body, {
			status: resp.status,
			statusText: resp.statusText,
			headers: resp.headers,
		});
	}

	return resolve(event);
}
