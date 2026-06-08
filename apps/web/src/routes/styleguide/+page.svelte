<script>
	import Button from '$lib/components/Button.svelte';
	import Input from '$lib/components/Input.svelte';
	import StatusBadge from '$lib/components/StatusBadge.svelte';

	let demo = $state('');
</script>

<header class="head">
	<p class="eyebrow">Lumni Design System</p>
	<h1>Styleguide</h1>
	<p class="muted">Tokens e componentes base consumidos apenas por role tokens.</p>
</header>

<section>
	<h2>Tipografia</h2>
	<div class="panel">
		<p class="t-display">Display 84</p>
		<p class="t-h1">Heading 1</p>
		<p class="t-h3">Heading 3</p>
		<p class="t-body">Body — o texto funcional padrão, em Figtree, 15px, leading generoso.</p>
		<p class="t-serif">Um acento editorial em Instrument Serif.</p>
		<p class="t-mono">mono · 0042 · /var/run · 2026-06-07</p>
	</div>
</section>

<section>
	<h2>Botões</h2>
	<div class="panel row">
		<Button variant="primary">Primary</Button>
		<Button variant="secondary">Secondary</Button>
		<Button variant="ghost">Ghost</Button>
		<Button variant="primary" size="sm">Small</Button>
		<Button variant="primary" size="lg">Large</Button>
		<Button variant="primary" disabled>Disabled</Button>
	</div>
</section>

<section>
	<h2>Inputs</h2>
	<div class="panel" style="max-width: 320px">
		<Input label="Campo de exemplo" bind:value={demo} placeholder="Digite algo…" />
	</div>
</section>

<section>
	<h2>Status badges</h2>
	<div class="panel row">
		<StatusBadge status="success" label="No ar" />
		<StatusBadge status="warning" label="Degradado" />
		<StatusBadge status="danger" label="Fora do ar" />
		<StatusBadge status="info" label="Planejado" />
	</div>
</section>

<style>
	.head {
		margin-bottom: var(--space-8);
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
		font-size: var(--text-3xl);
		font-weight: var(--weight-medium);
		letter-spacing: var(--tracking-tight);
		color: var(--color-text);
		margin: 0 0 var(--space-2);
	}
	h2 {
		font-size: var(--text-lg);
		font-weight: var(--weight-medium);
		color: var(--color-text);
		margin: 0 0 var(--space-3);
	}
	section {
		margin-bottom: var(--space-8);
	}
	.muted {
		color: var(--color-text-secondary);
	}
	.panel {
		background-color: var(--color-surface);
		border: var(--border-width-1) solid var(--color-border);
		border-radius: var(--radius-lg);
		padding: var(--space-6);
		box-shadow: var(--shadow-sm);
	}
	.row {
		display: flex;
		flex-wrap: wrap;
		gap: var(--space-3);
		align-items: center;
	}
	.t-display {
		font-size: var(--text-display);
		font-weight: var(--weight-medium);
		letter-spacing: var(--tracking-display);
		line-height: var(--leading-display);
		color: var(--color-text);
		margin: 0 0 var(--space-3);
	}
	.t-h1 {
		font-size: var(--text-4xl);
		font-weight: var(--weight-medium);
		letter-spacing: var(--tracking-tight);
		color: var(--color-text);
		margin: 0 0 var(--space-2);
	}
	.t-h3 {
		font-size: var(--text-2xl);
		font-weight: var(--weight-medium);
		color: var(--color-text);
		margin: 0 0 var(--space-2);
	}
	.t-body {
		font-size: var(--text-base);
		color: var(--color-text-secondary);
		margin: 0 0 var(--space-2);
	}
	.t-serif {
		font-family: var(--font-serif);
		font-style: italic;
		font-size: var(--text-xl);
		color: var(--color-accent);
		margin: 0 0 var(--space-2);
	}
	.t-mono {
		font-family: var(--font-mono);
		font-size: var(--text-sm);
		color: var(--color-text-muted);
		margin: 0;
	}
</style>
