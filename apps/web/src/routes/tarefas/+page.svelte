<script>
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import Button from '$lib/components/Button.svelte';
	import Input from '$lib/components/Input.svelte';
	import Select from '$lib/components/Select.svelte';
	import StatusBadge from '$lib/components/StatusBadge.svelte';
	import { api } from '$lib/api.js';
	import {
		STATUS,
		STATUS_COLUMNS,
		PRIORIDADE,
		PRIORIDADE_OPTIONS,
		fmtPrazo,
		isOverdue
	} from '$lib/taskMeta.js';

	const today = new Date().toISOString().slice(0, 10);

	let tasks = $state([]);
	let projects = $state([]);
	let loading = $state(true);
	let error = $state('');
	let filterProject = $state($page.url.searchParams.get('project') ?? '');
	let showForm = $state(false);

	// Create form.
	let titulo = $state('');
	let prioridade = $state('media');
	let prazo = $state('');
	let formProject = $state('');
	let tagsText = $state('');
	let saving = $state(false);
	let formError = $state('');

	let byStatus = $derived(
		Object.fromEntries(STATUS_COLUMNS.map((s) => [s, tasks.filter((t) => t.status === s)]))
	);
	let projectName = $derived(Object.fromEntries(projects.map((p) => [p.id, p.nome])));

	let projectFilterOptions = $derived([
		{ value: '', label: 'Todos os projetos' },
		...projects.map((p) => ({ value: p.id, label: p.nome }))
	]);
	let projectFormOptions = $derived([
		{ value: '', label: '— Sem projeto —' },
		...projects.map((p) => ({ value: p.id, label: p.nome }))
	]);

	async function loadProjects() {
		try {
			const r = await api('/api/projects');
			projects = r.projects;
		} catch {
			// A failed project list only costs us the names/filter; tasks still load.
		}
	}

	async function loadTasks(pid) {
		loading = true;
		error = '';
		try {
			const path = pid ? `/api/tasks?project_id=${encodeURIComponent(pid)}` : '/api/tasks';
			const r = await api(path);
			tasks = r.tasks;
		} catch (e) {
			error = e.message;
		} finally {
			loading = false;
		}
	}

	onMount(loadProjects);

	$effect(() => {
		const pid = filterProject; // tracked: reloads when the filter changes
		loadTasks(pid);
	});

	function toggleForm() {
		showForm = !showForm;
		if (showForm) formProject = filterProject; // default new task to the active filter
	}

	function resetForm() {
		titulo = prazo = tagsText = '';
		prioridade = 'media';
		formProject = filterProject;
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
			await api('/api/tasks', {
				method: 'POST',
				body: {
					titulo,
					prioridade,
					prazo: prazo || null,
					project_id: formProject || null,
					tags
				}
			});
			showForm = false;
			resetForm();
			await loadTasks(filterProject);
		} catch (e) {
			formError = e.message;
		} finally {
			saving = false;
		}
	}

	async function move(task, delta) {
		const idx = STATUS_COLUMNS.indexOf(task.status);
		const next = STATUS_COLUMNS[idx + delta];
		if (!next) return;
		try {
			await api(`/api/tasks/${task.id}`, { method: 'PATCH', body: { status: next } });
			await loadTasks(filterProject);
		} catch (e) {
			error = e.message;
		}
	}

	async function remove(task) {
		if (!confirm(`Excluir a tarefa "${task.titulo}"?`)) return;
		try {
			await api(`/api/tasks/${task.id}`, { method: 'DELETE' });
			await loadTasks(filterProject);
		} catch (e) {
			error = e.message;
		}
	}
</script>

<header class="page-head">
	<div>
		<p class="eyebrow">Tarefas</p>
		<h1>Quadro</h1>
	</div>
	<div class="head-actions">
		<Select bind:value={filterProject} options={projectFilterOptions} />
		<Button onclick={toggleForm}>{showForm ? 'Cancelar' : 'Nova tarefa'}</Button>
	</div>
</header>

{#if showForm}
	<form class="panel form" onsubmit={create}>
		<div class="grid">
			<Input label="Título" bind:value={titulo} required />
			<Select label="Prioridade" bind:value={prioridade} options={PRIORIDADE_OPTIONS} />
			<Input label="Prazo" type="date" bind:value={prazo} />
			<Select label="Projeto" bind:value={formProject} options={projectFormOptions} />
			<Input label="Tags (vírgula)" bind:value={tagsText} placeholder="bug, urgente" />
		</div>
		{#if formError}<p class="error">{formError}</p>{/if}
		<div class="actions">
			<Button type="submit" disabled={saving}>{saving ? 'Salvando…' : 'Criar tarefa'}</Button>
		</div>
	</form>
{/if}

{#if error}<p class="error">{error}</p>{/if}

{#if loading}
	<p class="muted">Carregando…</p>
{:else}
	<div class="board">
		{#each STATUS_COLUMNS as col (col)}
			<section class="column">
				<header class="col-head">
					<StatusBadge status={STATUS[col].variant} label={STATUS[col].label} />
					<span class="count">{byStatus[col].length}</span>
				</header>
				<div class="cards">
					{#each byStatus[col] as t (t.id)}
						<article class="card">
							<p class="titulo">{t.titulo}</p>
							<div class="meta">
								<StatusBadge status={PRIORIDADE[t.prioridade].variant} label={PRIORIDADE[t.prioridade].label} />
								{#if t.prazo}
									<span class="prazo" class:overdue={isOverdue(t, today)}>⏷ {fmtPrazo(t.prazo)}</span>
								{/if}
								{#if t.project_id && projectName[t.project_id]}
									<span class="proj">{projectName[t.project_id]}</span>
								{/if}
							</div>
							{#if t.tags.length}
								<div class="tags">
									{#each t.tags as tag (tag)}<span class="tag">{tag}</span>{/each}
								</div>
							{/if}
							<div class="card-actions">
								<button
									class="nudge"
									onclick={() => move(t, -1)}
									disabled={col === STATUS_COLUMNS[0]}
									aria-label="Mover para a coluna anterior">←</button
								>
								<button
									class="nudge"
									onclick={() => move(t, 1)}
									disabled={col === STATUS_COLUMNS[STATUS_COLUMNS.length - 1]}
									aria-label="Mover para a próxima coluna">→</button
								>
								<button class="del" onclick={() => remove(t)} aria-label="Excluir tarefa">excluir</button>
							</div>
						</article>
					{/each}
					{#if byStatus[col].length === 0}
						<p class="col-empty">—</p>
					{/if}
				</div>
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
		flex-wrap: wrap;
	}
	.head-actions {
		display: flex;
		align-items: flex-end;
		gap: var(--space-3);
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
		margin-bottom: var(--space-4);
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
		grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
		gap: var(--space-4);
	}
	.actions {
		display: flex;
		justify-content: flex-end;
	}

	.board {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: var(--space-4);
		align-items: start;
	}
	.column {
		background-color: var(--vapor);
		border: var(--border-width-1) solid var(--color-border);
		border-radius: var(--radius-lg);
		padding: var(--space-3);
		min-height: 120px;
	}
	.col-head {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 2px 4px var(--space-3);
	}
	.count {
		font-family: var(--font-mono);
		font-size: 12px;
		color: var(--color-text-muted);
	}
	.cards {
		display: flex;
		flex-direction: column;
		gap: var(--space-3);
	}
	.col-empty {
		color: var(--color-text-disabled);
		text-align: center;
		font-size: var(--text-sm);
		margin: var(--space-4) 0;
	}
	.card {
		background-color: var(--color-surface);
		border: var(--border-width-1) solid var(--color-border);
		border-radius: var(--radius-md);
		box-shadow: var(--shadow-sm);
		padding: var(--space-3) var(--space-4);
		display: flex;
		flex-direction: column;
		gap: var(--space-2);
	}
	.titulo {
		margin: 0;
		font-size: 13.5px;
		font-weight: var(--weight-semibold);
		color: var(--color-text);
	}
	.meta {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: var(--space-2);
	}
	.prazo {
		font-family: var(--font-mono);
		font-size: 11px;
		color: var(--color-text-muted);
	}
	.prazo.overdue {
		color: var(--color-danger-text);
		font-weight: var(--weight-semibold);
	}
	.proj {
		font-size: 11px;
		color: var(--color-text-secondary);
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
	.card-actions {
		display: flex;
		align-items: center;
		gap: var(--space-2);
		margin-top: var(--space-1);
	}
	.nudge {
		font-family: var(--font-mono);
		font-size: 13px;
		line-height: 1;
		padding: 4px 9px;
		background-color: var(--color-surface);
		border: var(--border-width-1) solid var(--color-border);
		border-radius: var(--radius-md);
		color: var(--color-text-secondary);
		cursor: pointer;
		transition:
			border-color var(--dur-fast) var(--ease-out),
			color var(--dur-fast) var(--ease-out);
	}
	.nudge:hover:not(:disabled) {
		border-color: var(--color-border-strong);
		color: var(--color-text);
	}
	.nudge:disabled {
		opacity: 0.35;
		cursor: not-allowed;
	}
	.del {
		margin-left: auto;
		background: none;
		border: none;
		color: var(--color-danger-text);
		font-size: var(--text-sm);
		cursor: pointer;
	}
	@media (max-width: 820px) {
		.board {
			grid-template-columns: 1fr;
		}
	}
</style>
