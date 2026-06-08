import { api } from '$lib/api.js';

// App-wide monitor state: SSE connection, alerts and the last live transition.
class MonitorStore {
	connected = $state(false);
	alerts = $state([]);
	unreadCount = $state(0);
	lastEvent = $state(null); // last monitor.transition payload (board reacts to it)

	setConnected(v) {
		this.connected = v;
	}

	onTransition(ev) {
		this.lastEvent = ev;
		this.unreadCount += 1;
		this.alerts = [
			{
				id: ev.alert_id,
				severity: ev.severity,
				title: titleFor(ev),
				to_status: ev.to,
				created_at: ev.at,
				read_at: null
			},
			...this.alerts
		];
	}

	async loadAlerts() {
		const res = await api('/api/alerts');
		this.alerts = res.alerts;
		this.unreadCount = res.unread_count;
	}

	async markAllRead() {
		await api('/api/alerts/read-all', { method: 'POST' });
		this.unreadCount = 0;
		this.alerts = this.alerts.map((a) => ({ ...a, read_at: a.read_at ?? new Date().toISOString() }));
	}
}

function titleFor(ev) {
	const m = { up: 'está no ar', degraded: 'está degradado', down: 'está fora do ar' };
	return `${ev.nome} ${m[ev.to] ?? 'mudou de estado'}`;
}

export const monitor = new MonitorStore();
