// Money helpers. Amounts are integers in the currency's minor unit (cents).
// Currencies are summed separately — there is no conversion (see ADR-0005).

const SYMBOL = { BRL: 'R$', USD: 'US$' };

export const MOEDA_OPTIONS = [
	{ value: 'BRL', label: 'R$ (BRL)' },
	{ value: 'USD', label: 'US$ (USD)' }
];

export const CICLO_OPTIONS = [
	{ value: 'mensal', label: 'Mensal' },
	{ value: 'anual', label: 'Anual' }
];

// Format cents in a currency, e.g. formatMoney(1900, 'USD') → "US$ 19,00".
export function formatMoney(cents, moeda = 'BRL') {
	const sym = SYMBOL[moeda] ?? moeda;
	const value = (cents / 100).toLocaleString('pt-BR', {
		minimumFractionDigits: 2,
		maximumFractionDigits: 2
	});
	return `${sym} ${value}`;
}

// Convert a major-unit string/number (e.g. "19.90") to integer cents.
export function toCents(amount) {
	const n = typeof amount === 'string' ? parseFloat(amount.replace(',', '.')) : Number(amount);
	if (!Number.isFinite(n) || n < 0) return 0;
	return Math.round(n * 100);
}

// Major-unit string for editing, e.g. centsToAmount(1900) → "19.00".
export function centsToAmount(cents) {
	return (cents / 100).toFixed(2);
}

// Monthly-normalized amount in cents (annual ÷ 12, rounded).
export function monthlyCents(sub) {
	return sub.ciclo === 'anual' ? Math.round(sub.valor_cents / 12) : sub.valor_cents;
}

// Sum active subscriptions' monthly cost grouped by currency → { BRL, USD } in cents.
export function sumByCurrency(subs, { onlyActive = true } = {}) {
	const totals = {};
	for (const s of subs) {
		if (onlyActive && !s.ativo) continue;
		totals[s.moeda] = (totals[s.moeda] ?? 0) + monthlyCents(s);
	}
	return totals;
}

// Render a totals map as "R$ 40,00/mês + US$ 25,00/mês" (or "—" if empty).
export function formatMonthlyTotals(totals) {
	const parts = Object.entries(totals)
		.filter(([, cents]) => cents > 0)
		.map(([moeda, cents]) => `${formatMoney(cents, moeda)}/mês`);
	return parts.length ? parts.join(' + ') : '—';
}
