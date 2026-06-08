// Maps project status → StatusBadge variant + label.
export const STATUS = {
	ideia: { label: 'Ideia', variant: 'info' },
	ativo: { label: 'Ativo', variant: 'success' },
	pausado: { label: 'Pausado', variant: 'warning' },
	no_ar: { label: 'No ar', variant: 'success' },
	arquivado: { label: 'Arquivado', variant: 'info' }
};

export const STATUS_OPTIONS = Object.entries(STATUS).map(([value, v]) => ({ value, label: v.label }));

export const VIS_OPTIONS = [
	{ value: 'pessoal', label: 'Pessoal' },
	{ value: 'lumni', label: 'Lumni' },
	{ value: 'cliente', label: 'Cliente' }
];

export const LINK_KINDS = [
	{ value: 'prod', label: 'Produção' },
	{ value: 'staging', label: 'Staging' },
	{ value: 'repo', label: 'Repositório' },
	{ value: 'docs', label: 'Docs' },
	{ value: 'design', label: 'Design' },
	{ value: 'other', label: 'Outro' }
];
