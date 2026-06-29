// Proxy all /api/* and /healthz requests to the Go API at runtime.
// API_URL is set as a Railway variable (or env var) and read here server-side.
// This replaces Vite's dev-only proxy for production (adapter-node).

const API_URL = process.env.API_URL || 'http://localhost:8080';

export async function handle({ event, resolve }) {
	const { pathname } = event.url;

	if (pathname.startsWith('/api') || pathname === '/healthz') {
		const upstream = API_URL + event.url.pathname + event.url.search;

		// Ask the upstream for an uncompressed body. Node's fetch (undici)
		// transparently decodes a compressed response, but the original
		// Content-Encoding header would then be forwarded alongside the already
		// decoded body, making the browser fail with ERR_CONTENT_DECODING_FAILED.
		// Stripping Accept-Encoding here keeps the hop identity-encoded; the web's
		// own edge still compresses the final response to the browser.
		const reqHeaders = new Headers(event.request.headers);
		reqHeaders.delete('accept-encoding');

		/** @type {RequestInit} */
		const init = {
			method: event.request.method,
			headers: reqHeaders,
		};

		if (event.request.method !== 'GET' && event.request.method !== 'HEAD') {
			init.body = await event.request.arrayBuffer();
		}

		const resp = await fetch(upstream, init);

		// Defensive: never forward an encoding/length that wouldn't match the body
		// we actually pass through (in case the upstream compressed anyway).
		const resHeaders = new Headers(resp.headers);
		resHeaders.delete('content-encoding');
		resHeaders.delete('content-length');

		return new Response(resp.body, {
			status: resp.status,
			statusText: resp.statusText,
			headers: resHeaders,
		});
	}

	return resolve(event);
}
