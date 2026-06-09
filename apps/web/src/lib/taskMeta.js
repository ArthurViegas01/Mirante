// Maps task status/priority → StatusBadge variant + label, plus the kanban
// column order and small prazo (due-date) helpers. Mirrors projectStatus.js.

export const STATUS = {
	a_fazer: { label: 'A fazer', variant: 'info' },
	fazendo: { label: 'Fazendo', variant: 'warning' },
	feito: { label: 'Feito', variant: 'success' }
};

// Kanban columns, left → right.
export const COLUMNS = ['a_fazer', 'fazendo', 'feito'];

export const STATUS_OPTIONS = Object.entries(STATUS).map(([value, v]) => ({ value, label: v.label }));

export const PRIORITY = {
	baixa: { label: 'Baixa', variant: 'info' },
	media: { label: 'Média', variant: 'warning' },
	alta: { label: 'Alta', variant: 'danger' }
};

export const PRIORITY_OPTIONS = Object.entries(PRIORITY).map(([value, v]) => ({ value, label: v.label }));

// --- prazo (due date) helpers; prazo is a 'YYYY-MM-DD' calendar date ---

// Parse 'YYYY-MM-DD' into a local Date at midnight (avoids the UTC-shift that
// `new Date('YYYY-MM-DD')` introduces).
function parseDate(s) {
	if (!s) return null;
	const [y, m, d] = s.split('-').map(Number);
	if (!y || !m || !d) return null;
	return new Date(y, m - 1, d);
}

// Whole days from today to the due date (negative = overdue); null if no prazo.
export function daysUntil(prazo) {
	const due = parseDate(prazo);
	if (!due) return null;
	const today = new Date();
	today.setHours(0, 0, 0, 0);
	return Math.round((due - today) / 86400000);
}

// A task is overdue when its prazo is in the past and it is not yet done.
export function isOverdue(prazo, status) {
	if (status === 'feito') return false;
	const d = daysUntil(prazo);
	return d !== null && d < 0;
}

// Human label for a prazo relative to today.
export function prazoLabel(prazo) {
	const d = daysUntil(prazo);
	if (d === null) return '';
	if (d < 0) return `Atrasada ${-d}d`;
	if (d === 0) return 'Vence hoje';
	if (d === 1) return 'Vence amanhã';
	return `Vence em ${d}d`;
}
