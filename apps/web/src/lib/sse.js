import { monitor } from '$lib/stores/monitor.svelte.js';

let es = null;

// connectMonitorStream opens the SSE stream (cookie auth, same-origin via the
// dev proxy). EventSource auto-reconnects and replays via Last-Event-ID.
export function connectMonitorStream() {
	if (es) return;
	es = new EventSource('/api/stream/monitor');
	es.onopen = () => monitor.setConnected(true);
	es.onerror = () => monitor.setConnected(false);
	es.addEventListener('monitor.transition', (e) => {
		try {
			monitor.onTransition(JSON.parse(e.data));
		} catch (_) {
			/* ignore malformed frame */
		}
	});
}

export function disconnectMonitorStream() {
	if (es) {
		es.close();
		es = null;
		monitor.setConnected(false);
	}
}
