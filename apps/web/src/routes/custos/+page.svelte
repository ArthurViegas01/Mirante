<script>
	import { onMount } from 'svelte';
	import { api } from '$lib/api.js';
	import {
		formatMoney,
		sumByCurrency,
		formatMonthlyTotals
	} from '$lib/money.js';

	let subscriptions = $state([]);
	let projects = $state([]);
	let loading = $state(true);
	let error = $state('');

	let grandTotal = $derived(formatMonthlyTotals(sumByCurrency(subscriptions)));

	// One group per project that has subscriptions, in project order.
	let groups = $derived.by(() => {
		const byProj = {};
		for (const s of subscriptions) (byProj[s.project_id] ??= []).push(s);
		return projects
			.filter((p) => byProj[p.id]?.length)
			.map((p) => ({
				project: p,
				subs: byProj[p.id],
				subtotal: formatMonthlyTotals(sumByCurrency(byProj[p.id]))
			}));
	});

	async function load() {
		loading = true;
		error = '';
		try {
			const [subs, proj] = await Promise.all([api('/api/subscriptions'), api('/api/projects')]);
			subscriptions = subs.subscriptions;
			projects = proj.projects;
		} catch (e) {
			error = e.message;
		} finally {
			loading = false;
		}
	}

	onMount(load);
</script>

<header class="page-head">
	<div>
		<p class="eyebrow">Custos</p>
		<h1>Assinaturas</h1>
	</div>
	{#if !loading && subscriptions.length}
		<div class="grand">
			<span class="grand-val">{grandTotal}</span>
			<span class="grand-cap">estimativa mensal</span>
		</div>
	{/if}
</header>

{#if loading}
	<p class="muted">Carregando…</p>
{:else if error}
	<p class="error">{error}</p>
{:else if subscriptions.length === 0}
	<div class="panel empty">
		Nenhuma assinatura registrada. Adicione custos na seção <strong>Custos</strong> de cada projeto.
	</div>
{:else}
	<div class="groups">
		{#each groups as g (g.project.id)}
			<section class="panel">
				<div class="group-head">
					<a class="proj" href={`/projetos/${g.project.id}`}>{g.project.nome}</a>
					<span class="subtotal">{g.subtotal} / mês</span>
				</div>
				<ul class="subs">
					{#each g.subs as s (s.id)}
						<li class:inactive={!s.ativo}>
							<span class="sub-nome">
								{s.nome}{#if s.provider}<span class="sub-prov"> · {s.provider}</span>{/if}{#if !s.ativo}<span class="sub-prov"> · pausada</span>{/if}
							</span>
							<span class="sub-valor">{formatMoney(s.valor_cents, s.moeda)} / {s.ciclo === 'anual' ? 'ano' : 'mês'}</span>
						</li>
					{/each}
				</ul>
			</section>
		{/each}
	</div>
{/if}

<style>
	.page-head {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: var(--space-4);
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
	.grand {
		display: flex;
		flex-direction: column;
		align-items: flex-end;
	}
	.grand-val {
		font-size: var(--text-lg);
		font-weight: var(--weight-semibold);
		color: var(--color-text);
	}
	.grand-cap {
		font-family: var(--font-mono);
		font-size: 11px;
		text-transform: uppercase;
		letter-spacing: 0.06em;
		color: var(--color-text-muted);
	}
	.muted {
		color: var(--color-text-secondary);
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
		padding: var(--space-8);
		text-align: center;
		color: var(--color-text-muted);
	}
	.groups {
		display: flex;
		flex-direction: column;
		gap: var(--space-4);
	}
	.group-head {
		display: flex;
		align-items: baseline;
		justify-content: space-between;
		gap: var(--space-4);
		padding: var(--space-4) var(--space-6);
		border-bottom: var(--border-width-1) solid var(--color-divider);
	}
	.proj {
		font-weight: var(--weight-semibold);
		color: var(--color-text);
		text-decoration: none;
	}
	.proj:hover {
		color: var(--color-link);
	}
	.subtotal {
		font-family: var(--font-mono);
		font-size: 13px;
		color: var(--color-text-secondary);
	}
	.subs {
		list-style: none;
		margin: 0;
		padding: var(--space-3) var(--space-6) var(--space-4);
		display: flex;
		flex-direction: column;
		gap: var(--space-2);
	}
	.subs li {
		display: flex;
		align-items: baseline;
		justify-content: space-between;
		gap: var(--space-3);
		font-size: var(--text-sm);
	}
	.subs li.inactive {
		opacity: 0.55;
	}
	.sub-nome {
		color: var(--color-text);
	}
	.sub-prov {
		font-family: var(--font-mono);
		font-size: 11px;
		color: var(--color-text-muted);
	}
	.sub-valor {
		font-family: var(--font-mono);
		font-size: 13px;
		color: var(--color-text);
		white-space: nowrap;
	}
</style>
