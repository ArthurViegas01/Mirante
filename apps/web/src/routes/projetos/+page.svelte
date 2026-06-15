<script>
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import Button from '$lib/components/Button.svelte';
	import Input from '$lib/components/Input.svelte';
	import Textarea from '$lib/components/Textarea.svelte';
	import Select from '$lib/components/Select.svelte';
	import StatusBadge from '$lib/components/StatusBadge.svelte';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import Skeleton from '$lib/components/Skeleton.svelte';
	import { api } from '$lib/api.js';
	import { toasts } from '$lib/stores/toast.svelte.js';
	import { STATUS, STATUS_OPTIONS, VIS_OPTIONS } from '$lib/projectStatus.js';

	let projects = $state([]);
	let loading = $state(true);
	let error = $state('');
	let showForm = $state(false);

	let nome = $state('');
	let codinome = $state('');
	let descricao = $state('');
	let repo = $state('');
	let status = $state('ideia');
	let visibilidade = $state('pessoal');
	let tagsText = $state('');
	let saving = $state(false);
	let formError = $state('');

	async function load() {
		loading = true;
		error = '';
		try {
			const res = await api('/api/projects');
			projects = res.projects;
		} catch (e) {
			error = e.message;
			toasts.error(e.message);
		} finally {
			loading = false;
		}
	}

	onMount(load);

	function resetForm() {
		nome = codinome = descricao = repo = tagsText = '';
		status = 'ideia';
		visibilidade = 'pessoal';
		formError = '';
	}

	async function create(e) {
		e.preventDefault();
		saving = true;
		formError = '';
		try {
			const tags = tagsText
				.split(',')
				.map((s) => s.trim())
				.filter(Boolean);
			await api('/api/projects', {
				method: 'POST',
				body: { nome, codinome, descricao, repo, status, visibilidade, tags }
			});
			showForm = false;
			resetForm();
			toasts.success('Projeto criado');
			await load();
		} catch (e) {
			formError = e.message;
			toasts.error(e.message);
		} finally {
			saving = false;
		}
	}
</script>

<header class="page-head">
	<div>
		<p class="eyebrow">Projetos</p>
		<h1>Projetos</h1>
	</div>
	<Button onclick={() => (showForm = !showForm)}>{showForm ? 'Cancelar' : 'Novo projeto'}</Button>
</header>

{#if showForm}
	<form class="panel form" onsubmit={create}>
		<div class="grid">
			<Input label="Nome" bind:value={nome} required />
			<Input label="Codinome" bind:value={codinome} />
			<Select label="Status" bind:value={status} options={STATUS_OPTIONS} />
			<Select label="Visibilidade" bind:value={visibilidade} options={VIS_OPTIONS} />
			<Input label="Repo" bind:value={repo} placeholder="https://github.com/…" />
			<Input label="Tags (vírgula)" bind:value={tagsText} placeholder="Go, SvelteKit" />
		</div>
		<Textarea label="Descrição" bind:value={descricao} rows={3} placeholder="O que é este projeto?" />
		{#if formError}<p class="error">{formError}</p>{/if}
		<div class="actions">
			<Button type="submit" disabled={saving}>{saving ? 'Salvando…' : 'Criar projeto'}</Button>
		</div>
	</form>
{/if}

{#if loading}
	<div class="table-wrap" aria-hidden="true">
		<table>
			<thead>
				<tr>
					<th>Projeto</th>
					<th>Status</th>
					<th>Stack</th>
					<th>Visibilidade</th>
				</tr>
			</thead>
			<tbody>
				{#each Array(5) as _, i (i)}
					<tr class="sk-row">
						<td class="name">
							<Skeleton w="55%" h="14px" />
							<Skeleton w="35%" h="11px" radius="var(--radius-sm)" block />
						</td>
						<td><Skeleton w="68px" h="20px" radius="var(--radius-full)" /></td>
						<td><Skeleton w="80%" h="12px" /></td>
						<td><Skeleton w="50%" h="12px" /></td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>
{:else if error}
	<p class="error">{error}</p>
{:else if projects.length === 0}
	<EmptyState
		title="Nenhum projeto ainda"
		description="Crie o primeiro projeto para começar a acompanhar status, links e tarefas."
	>
		{#snippet children()}
			<Button onclick={() => (showForm = true)}>Novo projeto</Button>
		{/snippet}
	</EmptyState>
{:else}
	<div class="table-wrap">
		<table>
			<thead>
				<tr>
					<th>Projeto</th>
					<th>Status</th>
					<th>Stack</th>
					<th>Visibilidade</th>
				</tr>
			</thead>
			<tbody>
				{#each projects as p (p.id)}
					<tr onclick={() => goto(`/projetos/${p.id}`)}>
						<td class="name">
							<span class="primary">{p.nome}</span>
							{#if p.codinome}<span class="meta">{p.codinome}</span>{/if}
						</td>
						<td>
							<StatusBadge
								status={STATUS[p.status]?.variant ?? 'info'}
								label={STATUS[p.status]?.label ?? p.status}
							/>
						</td>
						<td class="tags">
							{#each p.tags as t (t)}<span class="tag">{t}</span>{/each}
						</td>
						<td class="vis">{p.visibilidade}</td>
					</tr>
				{/each}
			</tbody>
		</table>
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
	.form {
		padding: var(--space-6);
		margin-bottom: var(--space-6);
		display: flex;
		flex-direction: column;
		gap: var(--space-4);
	}
	.grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: var(--space-4);
	}
	.actions {
		display: flex;
		justify-content: flex-end;
	}
	.sk-row td {
		cursor: default;
	}
	.sk-row .name {
		display: flex;
		flex-direction: column;
		gap: 6px;
	}
	.table-wrap {
		background-color: var(--color-surface);
		border: var(--border-width-1) solid var(--color-border);
		border-radius: var(--radius-lg);
		overflow: hidden;
	}
	table {
		width: 100%;
		border-collapse: collapse;
	}
	th {
		text-align: left;
		background-color: var(--color-surface-sunken);
		font-family: var(--font-mono);
		font-size: 11px;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		color: var(--color-text-muted);
		font-weight: var(--weight-medium);
		padding: 11px 18px;
		border-bottom: var(--border-width-1) solid var(--color-border);
	}
	td {
		padding: 14px 18px;
		font-size: 13.5px;
		color: var(--color-text);
		border-bottom: var(--border-width-1) solid var(--color-divider);
	}
	tbody tr {
		cursor: pointer;
		transition: background-color var(--dur-fast) var(--ease-out);
	}
	tbody tr:hover {
		background-color: var(--color-surface-sunken);
	}
	tbody tr:last-child td {
		border-bottom: none;
	}
	.name .primary {
		font-weight: var(--weight-semibold);
	}
	.name .meta {
		display: block;
		font-family: var(--font-mono);
		font-size: 12px;
		color: var(--color-text-muted);
		margin-top: 2px;
	}
	.tags {
		display: flex;
		flex-wrap: wrap;
		gap: 4px;
	}
	.tag {
		font-family: var(--font-mono);
		font-size: 11px;
		padding: 2px 7px;
		border-radius: var(--radius-full);
		background-color: var(--color-divider);
		color: var(--color-text-secondary);
	}
	.vis {
		color: var(--color-text-muted);
		text-transform: capitalize;
	}
</style>
