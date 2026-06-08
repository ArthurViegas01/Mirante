// Thin fetch wrapper: always sends the session cookie, attaches the CSRF token
// on unsafe methods, and throws the server's error message on non-2xx.

let csrf = '';

export function setCsrf(token) {
	csrf = token || '';
}

export function getCsrf() {
	return csrf;
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
	const data = text ? JSON.parse(text) : null;

	if (!res.ok) {
		const message = data?.error?.message || res.statusText || 'request failed';
		throw new Error(message);
	}
	return data;
}
