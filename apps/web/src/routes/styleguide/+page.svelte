<script>
	import Button from '$lib/components/Button.svelte';
	import Input from '$lib/components/Input.svelte';
	import Textarea from '$lib/components/Textarea.svelte';
	import StatusBadge from '$lib/components/StatusBadge.svelte';
	import StatCard from '$lib/components/StatCard.svelte';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import Skeleton from '$lib/components/Skeleton.svelte';
	import Modal from '$lib/components/Modal.svelte';
	import { toasts } from '$lib/stores/toast.svelte.js';
	import { confirm } from '$lib/stores/confirm.svelte.js';

	let demo = $state('');
	let demoErr = $state('senha123');
	let demoArea = $state('');
	let modalOpen = $state(false);

	async function askDemo() {
		const ok = await confirm.ask({
			title: 'Confirmar ação?',
			message: 'Isto é só uma demonstração do diálogo de confirmação.',
			confirmLabel: 'Confirmar'
		});
		toasts.info(ok ? 'Você confirmou.' : 'Você cancelou.');
	}
</script>

<header class="head">
	<p class="eyebrow">Lumni Design System</p>
	<h1>Styleguide</h1>
	<p class="muted">Tokens e componentes base, consumidos apenas por role tokens.</p>
</header>

<section>
	<h2>Tipografia</h2>
	<div class="panel">
		<p class="t-display">Display 84</p>
		<p class="t-h1">Heading 1</p>
		<p class="t-h3">Heading 3</p>
		<p class="t-body">Body: o texto funcional padrão, em Figtree, 15px, leading generoso.</p>
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
		<Button variant="danger">Danger</Button>
		<Button variant="primary" size="sm">Small</Button>
		<Button variant="primary" size="lg">Large</Button>
		<Button variant="primary" disabled>Disabled</Button>
	</div>
</section>

<section>
	<h2>Campos</h2>
	<div class="panel fields">
		<Input label="Texto" bind:value={demo} placeholder="Digite algo…" />
		<Input label="Senha" type="password" bind:value={demo} placeholder="Mostrar/ocultar" />
		<Input label="Com erro" bind:value={demoErr} error="Este valor é inválido." />
		<Textarea label="Área de texto" bind:value={demoArea} placeholder="Várias linhas…" rows={3} />
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

<section>
	<h2>Stat cards</h2>
	<div class="cards">
		<StatCard label="Projetos ativos" value="4" hint="6 no total" href="/projetos" />
		<StatCard label="Serviços no ar" value="3/3" live tone="success" hint="tudo no ar" />
		<StatCard label="Tarefas abertas" value="7" tone="danger" hint="2 atrasadas" />
	</div>
</section>

<section>
	<h2>Feedback</h2>
	<div class="panel row">
		<Button variant="secondary" onclick={() => toasts.success('Tudo certo por aqui.')}>Toast sucesso</Button>
		<Button variant="secondary" onclick={() => toasts.error('Algo deu errado.')}>Toast erro</Button>
		<Button variant="secondary" onclick={() => toasts.info('Apenas um aviso.')}>Toast info</Button>
		<Button variant="secondary" onclick={askDemo}>Abrir confirmação</Button>
		<Button variant="secondary" onclick={() => (modalOpen = true)}>Abrir modal</Button>
	</div>
</section>

<section>
	<h2>Estados</h2>
	<div class="two">
		<div class="panel">
			<EmptyState title="Nada por aqui" description="Um estado vazio com ícone, título e descrição.">
				{#snippet children()}<Button size="sm">Criar o primeiro</Button>{/snippet}
			</EmptyState>
		</div>
		<div class="panel sk">
			<Skeleton w="40%" h="14px" />
			<Skeleton w="100%" h="34px" radius="var(--radius-md)" />
			<Skeleton w="80%" h="34px" radius="var(--radius-md)" />
			<Skeleton w="60%" h="34px" radius="var(--radius-md)" />
		</div>
	</div>
</section>

<Modal bind:open={modalOpen} title="Exemplo de modal">
	<p class="m-text">
		Modais aparecem com fade e leve escala, prendem o foco e fecham no Escape ou clicando fora.
	</p>
	{#snippet footer()}
		<Button variant="secondary" onclick={() => (modalOpen = false)}>Fechar</Button>
		<Button onclick={() => { modalOpen = false; toasts.success('Confirmado no modal.'); }}>Confirmar</Button>
	{/snippet}
</Modal>

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
	.fields {
		display: flex;
		flex-direction: column;
		gap: var(--space-4);
		max-width: 360px;
	}
	.cards {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
		gap: var(--space-4);
	}
	.two {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
		gap: var(--space-4);
	}
	.sk {
		display: flex;
		flex-direction: column;
		gap: var(--space-3);
	}
	.m-text {
		margin: 0;
		color: var(--color-text-secondary);
		line-height: var(--leading-relaxed);
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
