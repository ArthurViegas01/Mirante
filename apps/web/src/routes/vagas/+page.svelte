<script>
	import StatusBadge from '$lib/components/StatusBadge.svelte';
</script>

<header class="page-head">
	<div>
		<p class="eyebrow">Vagas &amp; CV</p>
		<h1>Vagas &amp; CV</h1>
	</div>
	<StatusBadge status="info" label="Em breve — F3" />
</header>
<p class="muted">
	Análise de aderência, CV mestre adaptável, e-mail/carta e o CRM de candidaturas vão morar aqui.
</p>

<style>
	.page-head {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: var(--space-4);
		margin-bottom: var(--space-4);
	}
	.eyebrow {
		font-family: var(--font-mono);
		font-size: var(--text-xs);
		letter-spacing: var(--tracking-eyebrow);
		text-transform: uppercase;
		color: var(--color-text-muted);
		margin: 0 0 var(--space-2);
	}
	h1 {
		font-size: var(--text-2xl);
		font-weight: var(--weight-medium);
		letter-spacing: var(--tracking-snug);
		color: var(--color-text);
		margin: 0;
	}
	.muted {
		color: var(--color-text-secondary);
		font-size: var(--text-base);
	}
</style>
