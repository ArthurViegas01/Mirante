<script>
	import { onMount } from 'svelte';
	import Button from '$lib/components/Button.svelte';
	import Input from '$lib/components/Input.svelte';
	import StatusBadge from '$lib/components/StatusBadge.svelte';
	import { api } from '$lib/api.js';
	import {
		APP_PIPELINE,
		APP_STATUS_OPTIONS,
		appStatusLabel,
		appStatusVariant
	} from '$lib/applicationMeta.js';

	let apps = $state([]);
	let loading = $state(true);
	let error = $state('');

	let editingId = $state('');
	let proximaAcao = $state('');
	let dataAcao = $state('');
	let notas = $state('');

	let groups = $derived(
		APP_PIPELINE.map((status) => ({
			status,
			items: apps.filter((a) => a.status === status)
		})).filter((g) => g.items.length)
	);

	async function load() {
		loading = true;
		error = '';
		try {
			const res = await api('/api/applications');
			apps = res.applications;
		} catch (e) {
			error = e.message;
		} finally {
			loading = false;
		}
	}

	onMount(load);

	async function changeStatus(app, status) {
		try {
			await api(`/api/applications/${app.id}`, { method: 'PATCH', body: { status } });
			await load();
		} catch (e) {
			error = e.message;
		}
	}

	function openEdit(app) {
		editingId = app.id;
		proximaAcao = app.proxima_acao ?? '';
		dataAcao = app.data_acao ?? '';
		notas = app.notas ?? '';
	}

	async function saveEdit(app) {
		try {
			await api(`/api/applications/${app.id}`, {
				method: 'PATCH',
				body: { proxima_acao: proximaAcao, data_acao: dataAcao, notas }
			});
			editingId = '';
			await load();
		} catch (e) {
			error = e.message;
		}
	}

	async function remove(app) {
		if (!confirm(`Remover a candidatura "${app.titulo}"?`)) return;
		try {
			await api(`/api/applications/${app.id}`, { method: 'DELETE' });
			await load();
		} catch (e) {
			error = e.message;
		}
	}
</script>

<header class="page-head">
	<div>
		<p class="eyebrow">Candidaturas</p>
		<h1>Pipeline</h1>
	</div>
</header>

{#if loading}
	<p class="muted">Carregando…</p>
{:else if error}
	<p class="error">{error}</p>
{:else if apps.length === 0}
	<div class="panel empty">
		Nenhuma candidatura ainda. Em <a href="/vagas">Vagas</a>, clique em "Acompanhar" para começar o
		pipeline.
	</div>
{:else}
	<div class="summary">
		{#each APP_PIPELINE as s (s)}
			{@const n = apps.filter((a) => a.status === s).length}
			<span class="stat" class:zero={n === 0}>
				<StatusBadge status={appStatusVariant(s)} label={appStatusLabel(s)} />
				<span class="n">{n}</span>
			</span>
		{/each}
	</div>

	<div class="groups">
		{#each groups as g (g.status)}
			<section class="group">
				<p class="group-h">{appStatusLabel(g.status)} · {g.items.length}</p>
				<div class="cards">
					{#each g.items as app (app.id)}
						<article class="card">
							<div class="card-top">
								<div class="who">
									<span class="titulo">{app.titulo}</span>
									{#if app.empresa}<span class="empresa">{app.empresa}</span>{/if}
								</div>
								<select
									class="status"
									value={app.status}
									onchange={(e) => changeStatus(app, e.currentTarget.value)}
									aria-label="Status"
								>
									{#each APP_STATUS_OPTIONS as o (o.value)}<option value={o.value}>{o.label}</option>{/each}
								</select>
							</div>

							{#if editingId === app.id}
								<div class="edit">
									<div class="edit-grid">
										<Input label="Próxima ação" bind:value={proximaAcao} placeholder="Enviar follow-up" />
										<Input label="Data" type="date" bind:value={dataAcao} />
									</div>
									<label class="field">
										<span class="label">Notas</span>
										<textarea bind:value={notas} rows="3"></textarea>
									</label>
									<div class="edit-actions">
										<Button size="sm" onclick={() => saveEdit(app)}>Salvar</Button>
										<button type="button" class="link" onclick={() => (editingId = '')}>cancelar</button>
									</div>
								</div>
							{:else}
								{#if app.proxima_acao || app.data_acao}
									<p class="acao">📌 {app.proxima_acao}{#if app.data_acao} · {app.data_acao}{/if}</p>
								{/if}
								{#if app.notas}<p class="notas">{app.notas}</p>{/if}
								<div class="card-actions">
									{#if app.job_id}<a class="link" href="/vagas">ver vagas</a>{/if}
									<button type="button" class="link" onclick={() => openEdit(app)}>editar</button>
									<button type="button" class="link danger" onclick={() => remove(app)}>remover</button>
								</div>
							{/if}
						</article>
					{/each}
				</div>
			</section>
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
	.empty a {
		color: var(--color-link);
	}
	.summary {
		display: flex;
		flex-wrap: wrap;
		gap: var(--space-4);
		margin-bottom: var(--space-6);
	}
	.stat {
		display: inline-flex;
		align-items: center;
		gap: 6px;
	}
	.stat.zero {
		opacity: 0.45;
	}
	.stat .n {
		font-family: var(--font-mono);
		font-size: 13px;
		color: var(--color-text);
	}
	.groups {
		display: flex;
		flex-direction: column;
		gap: var(--space-5);
	}
	.group-h {
		font-family: var(--font-mono);
		font-size: 11px;
		letter-spacing: 0.06em;
		text-transform: uppercase;
		color: var(--color-text-muted);
		margin: 0 0 var(--space-3);
	}
	.cards {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
		gap: var(--space-3);
	}
	.card {
		background-color: var(--color-surface);
		border: var(--border-width-1) solid var(--color-border);
		border-radius: var(--radius-lg);
		box-shadow: var(--shadow-sm);
		padding: var(--space-4);
		display: flex;
		flex-direction: column;
		gap: var(--space-2);
	}
	.card-top {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: var(--space-3);
	}
	.who {
		display: flex;
		flex-direction: column;
		min-width: 0;
	}
	.titulo {
		font-weight: var(--weight-semibold);
		color: var(--color-text);
	}
	.empresa {
		font-size: var(--text-sm);
		color: var(--color-text-secondary);
	}
	.status {
		font-family: var(--font-sans);
		font-size: 12px;
		padding: 5px 8px;
		background-color: var(--color-surface);
		color: var(--color-text);
		border: var(--border-width-1) solid var(--color-border);
		border-radius: var(--radius-md);
		cursor: pointer;
		flex-shrink: 0;
	}
	.acao {
		margin: 0;
		font-size: var(--text-sm);
		color: var(--color-text);
	}
	.notas {
		margin: 0;
		font-size: var(--text-sm);
		color: var(--color-text-secondary);
	}
	.card-actions {
		display: flex;
		gap: var(--space-3);
		border-top: var(--border-width-1) solid var(--color-divider);
		padding-top: var(--space-2);
		margin-top: var(--space-1);
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
	.link.danger:hover {
		color: var(--color-danger-text);
	}
	.edit {
		display: flex;
		flex-direction: column;
		gap: var(--space-3);
	}
	.edit-grid {
		display: grid;
		grid-template-columns: 2fr 1fr;
		gap: var(--space-3);
	}
	.field {
		display: flex;
		flex-direction: column;
		gap: var(--space-1);
	}
	.label {
		font-size: 11.5px;
		font-weight: var(--weight-medium);
		color: var(--color-text-secondary);
	}
	textarea {
		font-family: var(--font-sans);
		font-size: 13px;
		padding: 8px 10px;
		background-color: var(--color-surface);
		color: var(--color-text);
		border: var(--border-width-1) solid var(--color-border);
		border-radius: var(--radius-md);
		resize: vertical;
	}
	textarea:focus {
		outline: none;
		border-color: var(--color-primary);
		box-shadow: var(--shadow-focus);
	}
	.edit-actions {
		display: flex;
		align-items: center;
		gap: var(--space-3);
	}
</style>
