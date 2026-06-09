// Maps candidatura status → label + StatusBadge variant, and the pipeline order.
export const APP_STATUS = {
	interesse: { label: 'Interesse', variant: 'info' },
	aplicado: { label: 'Aplicado', variant: 'info' },
	entrevista: { label: 'Entrevista', variant: 'warning' },
	oferta: { label: 'Oferta', variant: 'success' },
	aceito: { label: 'Aceito', variant: 'success' },
	rejeitado: { label: 'Rejeitado', variant: 'danger' }
};

// Pipeline order, left → right.
export const APP_PIPELINE = ['interesse', 'aplicado', 'entrevista', 'oferta', 'aceito', 'rejeitado'];

export const APP_STATUS_OPTIONS = APP_PIPELINE.map((v) => ({ value: v, label: APP_STATUS[v].label }));

export function appStatusLabel(s) {
	return APP_STATUS[s]?.label ?? s;
}
export function appStatusVariant(s) {
	return APP_STATUS[s]?.variant ?? 'info';
}
