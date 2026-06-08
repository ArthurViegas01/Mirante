<script>
	import Button from '$lib/components/Button.svelte';
</script>

<section class="hero">
	<p class="eyebrow">Mirante</p>
	<h1>Sua central de comando.</h1>
	<p class="lead">
		Projetos, monitor <em>ao vivo</em>, tarefas e busca de carreira — num só lugar.
	</p>
	<div class="actions">
		<a href="/projetos"><Button>Entrar no painel</Button></a>
		<a href="/styleguide"><Button variant="secondary">Ver o styleguide</Button></a>
	</div>
</section>

<style>
	.hero {
		max-width: var(--max-prose);
	}
	.eyebrow {
		font-family: var(--font-mono);
		font-size: var(--text-xs);
		letter-spacing: var(--tracking-eyebrow);
		text-transform: uppercase;
		color: var(--color-text-muted);
		margin: 0 0 var(--space-3);
	}
	h1 {
		font-size: var(--text-4xl);
		font-weight: var(--weight-medium);
		letter-spacing: var(--tracking-tight);
		line-height: var(--leading-tight);
		color: var(--color-text);
		margin: 0 0 var(--space-4);
	}
	.lead {
		font-size: var(--text-md);
		line-height: var(--leading-relaxed);
		color: var(--color-text-secondary);
		margin: 0 0 var(--space-8);
	}
	.lead em {
		font-family: var(--font-serif);
		font-style: italic;
		color: var(--color-accent);
	}
	.actions {
		display: flex;
		gap: var(--space-3);
	}
	.actions a {
		text-decoration: none;
	}
</style>
