// Deterministic aderência (match) between a vaga's required skills and the user's
// master skills (from the CV profile). Case-insensitive on canonical names. The
// LLM-refined, experience-aware score comes later; this is the always-on floor.

export function aderencia(jobSkills, mySkills) {
	const mine = new Set((mySkills || []).map((s) => s.toLowerCase()));
	const required = jobSkills || [];
	const matched = required.filter((s) => mine.has(s.toLowerCase()));
	const missing = required.filter((s) => !mine.has(s.toLowerCase()));
	const score = required.length ? Math.round((matched.length / required.length) * 100) : null;
	return { score, matched, missing };
}

// Maps a score to a StatusBadge variant.
export function aderenciaVariant(score) {
	if (score === null) return 'info';
	if (score >= 70) return 'success';
	if (score >= 40) return 'warning';
	return 'danger';
}
