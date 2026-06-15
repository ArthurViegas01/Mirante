// Thin fetch wrapper: always sends the session cookie, attaches the CSRF token
// on unsafe methods, and throws the server's error message on non-2xx.

let csrf = '';

export function setCsrf(token) {
	csrf = token || '';
}

export function getCsrf() {
	return csrf;
}

// Global 401 handler. The layout registers a callback so an expired/revoked
// session (a 401 on a protected route) drops the user back to /login instead of
// failing silently. Kept as a decoupled callback so api.js never imports the
// session store. The callback itself decides whether to act (it no-ops on the
// expected 401 from the initial /me probe, when no session is established).
let onUnauthorized = null;

export function setUnauthorizedHandler(fn) {
	onUnauthorized = fn;
}

export async function api(path, { method = 'GET', body } = {}) {
	const headers = {};
	if (body !== undefined) headers['Content-Type'] = 'application/json';
	if (method !== 'GET' && method !== 'HEAD' && csrf) headers['X-CSRF-Token'] = csrf;

	const res = await fetch(path, {
		method,
		credentials: 'include',
		headers,
		body: body !== undefined ? JSON.stringify(body) : undefined
	});

	const text = await res.text();
	let data = null;
	if (text) {
		try {
			data = JSON.parse(text);
		} catch {
			// Body wasn't JSON (an HTML error page, a proxy response, etc.). On an
			// error status, surface a clean message instead of a parse exception.
			if (!res.ok) throw new Error(res.statusText || 'request failed');
			data = text;
		}
	}

	if (!res.ok) {
		if (res.status === 401 && onUnauthorized) onUnauthorized();
		const message = data?.error?.message || res.statusText || 'request failed';
		throw new Error(message);
	}
	return data;
}
