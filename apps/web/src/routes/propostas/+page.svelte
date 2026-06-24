<script>
	import { onMount } from 'svelte';
	import StatusBadge from '$lib/components/StatusBadge.svelte';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import Skeleton from '$lib/components/Skeleton.svelte';
	import { api } from '$lib/api.js';
	import { toasts } from '$lib/stores/toast.svelte.js';
	import { confirm } from '$lib/stores/confirm.svelte.js';
	import { aderenciaVariant } from '$lib/aderencia.js';

	let items = $state([]);
	let mySkills = $state([]);
	let minScore = $state(60);
	let loading = $state(true);
	let error = $state('');
	let dismissingId = $state('');

	let mySet = $derived(new Set(mySkills.map((s) => s.toLowerCase())));
	let shortlist = $derived(items.filter((i) => i.score >= minScore).length);

	async function load() {
		loading = true;
		error = '';
		try {
			const [res, profile] = await Promise.all([
				api('/api/intake?estado=novo'),
				api('/api/profile')
			]);
			items = res.items ?? [];
			minScore = res.min_score ?? 60;
			mySkills = profile.skills ?? [];
		} catch (e) {
			error = e.message;
			toasts.error(e.message);
		} finally {
			loading = false;
		}
	}

	onMount(load);

	async function dismiss(item) {
		if (
			!(await confirm.ask({
				title: 'Descartar projeto?',
				message: `"${item.titulo}" sai da fila de triagem.`,
				confirmLabel: 'Descartar'
			}))
		)
			return;
		dismissingId = item.id;
		try {
			await api(`/api/intake/${item.id}/dismiss`, { method: 'POST' });
			items = items.filter((i) => i.id !== item.id);
			toasts.success('Projeto descartado');
		} catch (e) {
			toasts.error(e.message);
		} finally {
			dismissingId = '';
		}
	}

	function prazo(horas) {
		if (!horas || horas <= 0) return '';
		const d = Math.floor(horas / 24);
		const h = horas % 24;
		return d > 0 ? `${d}d ${h}h restantes` : `${h}h restantes`;
	}

	// Competition signal: fewer proposals is a hotter lead.
	function compVariant(n) {
		if (n <= 10) return 'success';
		if (n >= 40) return 'danger';
		return 'warning';
	}
</script>

<header class="page-head">
	<div>
		<p class="eyebrow">Propostas</p>
		<h1>Fila de triagem</h1>
		<p class="sub">Projetos do 99Freelas, ranqueados por aderência, concorrência e frescor.</p>
	</div>
</header>

{#if loading}
	<div class="list" aria-hidden="true">
		{#each Array(3) as _, i (i)}
			<article class="panel card sk">
				<div class="card-head">
					<Skeleton w="55%" h="18px" />
					<Skeleton w="44px" h="22px" radius="var(--radius-full)" />
				</div>
				<Skeleton w="40%" h="12px" />
				<Skeleton w="90%" h="12px" block />
				<div class="skills">
					{#each Array(4) as __, j (j)}<Skeleton w="58px" h="18px" radius="var(--radius-full)" />{/each}
				</div>
			</article>
		{/each}
	</div>
{:else if error}
	<p class="error">{error}</p>
{:else if items.length === 0}
	<div class="panel empty">
		<EmptyState
			title="Nenhum projeto na fila"
			description="Quando o poller do 99Freelas rodar, os projetos aparecem aqui — os melhores no topo. Configure as credenciais IMAP (INTAKE_IMAP_*) no .env para ligá-lo."
		/>
	</div>
{:else}
	<p class="count">
		<strong>{items.length}</strong> na fila · <strong>{shortlist}</strong> com score ≥ {minScore} (shortlist)
	</p>
	<div class="list">
		{#each items as item (item.id)}
			<article class="panel card" class:top={item.score >= minScore}>
				<div class="card-head">
					<div class="title-wrap">
						<h2>{item.titulo}</h2>
						<div class="meta">
							{#if item.categoria}<span>{item.categoria}</span>{/if}
							{#if item.nivel}<span>· {item.nivel}</span>{/if}
							{#if item.publicado}<span>· {item.publicado}</span>{/if}
						</div>
					</div>
					<div class="score" title="Score de triagem (0–100)">
						<StatusBadge status={aderenciaVariant(item.score)} label={`${item.score}`} />
					</div>
				</div>

				<div class="signals">
					<StatusBadge status={compVariant(item.propostas)} label={`${item.propostas} propostas`} />
					{#if item.interessados}<span class="sig">{item.interessados} interessados</span>{/if}
					{#if prazo(item.restante_horas)}<span class="sig">⏳ {prazo(item.restante_horas)}</span>{/if}
				</div>

				{#if item.teaser}<p class="teaser">{item.teaser}</p>{/if}

				{#if (item.skills ?? []).length}
					<div class="skills">
						{#each item.skills as s (s)}
							<span
								class="skill"
								class:have={mySet.has(s.toLowerCase())}
								class:lack={mySkills.length && !mySet.has(s.toLowerCase())}>{s}</span
							>
						{/each}
					</div>
				{/if}

				<div class="card-actions">
					{#if item.enviar_url}
						<a class="cta" href={item.enviar_url} target="_blank" rel="noreferrer">Enviar proposta ↗</a>
					{/if}
					{#if item.url}
						<a class="link" href={item.url} target="_blank" rel="noreferrer">ver projeto ↗</a>
					{/if}
					<button
						class="link danger"
						onclick={() => dismiss(item)}
						disabled={dismissingId === item.id}
					>
						{dismissingId === item.id ? '…' : 'descartar'}
					</button>
				</div>
			</article>
		{/each}
	</div>
{/if}

<style>
	.page-head {
		margin-bottom: var(--space-6);
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
	.sub {
		margin: var(--space-2) 0 0;
		font-size: var(--text-sm);
		color: var(--color-text-secondary);
		max-width: var(--max-prose);
	}
	.error {
		color: var(--color-danger-text);
		font-size: var(--text-sm);
	}
	.panel {
		background-color: var(--color-surface);
		border: var(--border-width-1) solid var(--color-border);
		border-radius: var(--radius-lg);
		box-shadow: var(--shadow-sm);
	}
	.empty {
		padding: var(--space-2);
	}
	.count {
		margin: 0 0 var(--space-4);
		font-size: var(--text-sm);
		color: var(--color-text-secondary);
	}
	.count strong {
		color: var(--color-text);
		font-variant-numeric: tabular-nums;
	}
	.list {
		display: flex;
		flex-direction: column;
		gap: var(--space-4);
	}
	.card {
		padding: var(--space-5) var(--space-6);
		display: flex;
		flex-direction: column;
		gap: var(--space-3);
	}
	/* The shortlist (score above the floor) gets the live Glow accent. */
	.card.top {
		border-color: color-mix(in srgb, var(--color-live) 40%, var(--color-border));
		box-shadow: 0 0 0 1px color-mix(in srgb, var(--color-live) 22%, transparent);
	}
	.card-head {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: var(--space-4);
	}
	.title-wrap {
		min-width: 0;
	}
	h2 {
		font-size: var(--text-lg);
		font-weight: var(--weight-semibold);
		color: var(--color-text);
		margin: 0;
	}
	.meta {
		display: flex;
		flex-wrap: wrap;
		gap: 4px;
		margin-top: 2px;
		font-size: var(--text-sm);
		color: var(--color-text-muted);
	}
	.score {
		flex-shrink: 0;
		font-variant-numeric: tabular-nums;
	}
	.signals {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: var(--space-3);
	}
	.sig {
		font-size: var(--text-sm);
		color: var(--color-text-secondary);
	}
	.teaser {
		margin: 0;
		font-family: var(--font-serif);
		font-style: italic;
		color: var(--color-text-secondary);
		max-width: var(--max-prose);
	}
	.skills {
		display: flex;
		flex-wrap: wrap;
		gap: 4px;
	}
	.skill {
		font-family: var(--font-mono);
		font-size: 11px;
		padding: 2px 7px;
		border-radius: var(--radius-full);
		background-color: color-mix(in srgb, var(--color-accent) 12%, transparent);
		color: var(--color-accent);
	}
	.skill.have {
		background-color: var(--color-success-bg);
		color: var(--color-success-text);
	}
	.skill.lack {
		background-color: transparent;
		color: var(--color-text-muted);
		border: var(--border-width-1) solid var(--color-border);
	}
	.card-actions {
		display: flex;
		align-items: center;
		gap: var(--space-4);
		border-top: var(--border-width-1) solid var(--color-divider);
		padding-top: var(--space-3);
	}
	.cta {
		font-size: var(--text-sm);
		font-weight: var(--weight-medium);
		color: var(--color-accent);
		text-decoration: none;
	}
	.cta:hover {
		text-decoration: underline;
	}
	.link {
		background: none;
		border: none;
		padding: 0;
		font-size: var(--text-sm);
		color: var(--color-text-muted);
		cursor: pointer;
		text-decoration: none;
	}
	.link:hover {
		color: var(--color-text);
	}
	.link.danger {
		margin-left: auto;
	}
	.link.danger:hover {
		color: var(--color-danger-text);
	}
</style>
