// Maps job "modelo" (work arrangement) to a label + select options.
export const MODELO = {
	remoto: 'Remoto',
	hibrido: 'Híbrido',
	presencial: 'Presencial',
	indefinido: 'Indefinido'
};

export const MODELO_OPTIONS = Object.entries(MODELO).map(([value, label]) => ({ value, label }));

export function modeloLabel(m) {
	return MODELO[m] ?? m;
}
