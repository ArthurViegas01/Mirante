// Maps task status/priority → StatusBadge variant + label, and the board layout.

export const STATUS = {
	a_fazer: { label: 'A fazer', variant: 'info' },
	fazendo: { label: 'Fazendo', variant: 'warning' },
	feito: { label: 'Feito', variant: 'success' }
};

// Board columns, left → right.
export const STATUS_COLUMNS = ['a_fazer', 'fazendo', 'feito'];

export const STATUS_OPTIONS = STATUS_COLUMNS.map((value) => ({ value, label: STATUS[value].label }));

export const PRIORIDADE = {
	baixa: { label: 'Baixa', variant: 'info' },
	media: { label: 'Média', variant: 'warning' },
	alta: { label: 'Alta', variant: 'danger' }
};

export const PRIORIDADE_OPTIONS = Object.entries(PRIORIDADE).map(([value, v]) => ({
	value,
	label: v.label
}));

// A task is "open" while it isn't done.
export function isOpen(task) {
	return task.status !== 'feito';
}

// prazo is an ISO date "YYYY-MM-DD"; show it as "dd/mm".
export function fmtPrazo(prazo) {
	if (!prazo) return '';
	const [, m, d] = prazo.split('-');
	return `${d}/${m}`;
}

// Overdue = has a past deadline and isn't done yet.
export function isOverdue(task, today) {
	return !!task.prazo && task.status !== 'feito' && task.prazo < today;
}
