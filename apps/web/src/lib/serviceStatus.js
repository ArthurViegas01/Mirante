// Maps monitor service status → StatusBadge variant + label.
export const SVC_STATUS = {
	unknown: { label: 'Desconhecido', variant: 'info' },
	up: { label: 'No ar', variant: 'success' },
	degraded: { label: 'Degradado', variant: 'warning' },
	down: { label: 'Fora do ar', variant: 'danger' },
	paused: { label: 'Pausado', variant: 'info' }
};

export function svcVariant(status) {
	return SVC_STATUS[status]?.variant ?? 'info';
}
export function svcLabel(status) {
	return SVC_STATUS[status]?.label ?? status;
}

export const KIND_OPTIONS = [
	{ value: 'http', label: 'HTTP' },
	{ value: 'tcp', label: 'TCP' },
	{ value: 'db_ping', label: 'Banco (ping)' }
];

export function uptimeVariant(ratio) {
	if (ratio >= 0.99) return 'success';
	if (ratio >= 0.9) return 'warning';
	return 'danger';
}
