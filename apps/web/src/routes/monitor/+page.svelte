<script>
	import StatusBadge from '$lib/components/StatusBadge.svelte';
</script>

<header class="page-head">
	<div>
		<p class="eyebrow">Monitor</p>
		<h1>Monitor</h1>
	</div>
	<StatusBadge status="info" label="Em breve — F1" />
</header>
<p class="muted">A saúde ao vivo dos serviços de cada projeto, com alertas in-app, aparece aqui.</p>

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
