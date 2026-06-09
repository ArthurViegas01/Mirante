<script>
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import Button from '$lib/components/Button.svelte';
	import Input from '$lib/components/Input.svelte';
	import Select from '$lib/components/Select.svelte';
	import StatusBadge from '$lib/components/StatusBadge.svelte';
	import { api } from '$lib/api.js';
	import { COLUMNS, STATUS, PRIORITY, PRIORITY_OPTIONS, isOverdue, prazoLabel } from '$lib/taskMeta.js';

	let tasks = $state([]);
	let projects = $state([]);
	let loading = $state(true);
	let error = $state('');
	let showForm = $state(false);

	// Active project filter, driven by the ?project= query param.
	let projectFilter = $derived($page.url.searchParams.get('project') ?? '');

	// Create-form state.
	let titulo = $state('');
	let descricao = $state('');
	let prioridade = $state('media');
	let prazo = $state('');
	let formProject = $state('');
	let tagsText = $state('');
	let saving = $state(false);
	let formError = $state('');

	let projectOptions = $derived([
		{ value: '', label: '— Sem projeto —' },
		...projects.map((p) => ({ value: p.id, label: p.nome }))
	]);

	const projectName = (id) => projects.find((p) => p.id === id)?.nome ?? 'projeto';
	const byStatus = (s) => tasks.filter((t) => t.status === s);

	async function load(filter) {
		loading = true;
		error = '';
		try {
			const q = filter ? `?project=${encodeURIComponent(filter)}` : '';
			const [tRes, pRes] = await Promise.all([api(`/api/tasks${q}`), api('/api/projects')]);
			tasks = tRes.tasks;
			projects = pRes.projects;
		} catch (e) {
			error = e.message;
		} finally {
			loading = false;
		}
	}

	// Reload whenever the project filter changes (initial mount included).
	$effect(() => {
		load(projectFilter);
	});

	function changeFilter(e) {
		const v = e.currentTarget.value;
		goto(v ? `/tarefas?project=${encodeURIComponent(v)}` : '/tarefas');
	}

	function openForm() {
		showForm = true;
		formProject = projectFilter;
	}

	function resetForm() {
		titulo = descricao = prazo = tagsText = '';
		prioridade = 'media';
		formProject = projectFilter;
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
					descricao,
					prioridade,
					prazo: prazo || '',
					project_id: formProject || '',
					tags
				}
			});
			showForm = false;
			resetForm();
			await load(projectFilter);
		} catch (e) {
			formError = e.message;
		} finally {
			saving = false;
		}
	}

	async function move(task, dir) {
		const next = COLUMNS[COLUMNS.indexOf(task.status) + dir];
		if (!next) return;
		try {
			await api(`/api/tasks/${task.id}`, { method: 'PATCH', body: { status: next } });
			await load(projectFilter);
		} catch (e) {
			error = e.message;
		}
	}

	async function remove(task) {
		if (!confirm('Excluir esta tarefa?')) return;
		try {
			await api(`/api/tasks/${task.id}`, { method: 'DELETE' });
			await load(projectFilter);
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
		<select
			class="filter"
			value={projectFilter}
			onchange={changeFilter}
			aria-label="Filtrar por projeto"
		>
			<option value="">Todos os projetos</option>
			{#each projects as p (p.id)}<option value={p.id}>{p.nome}</option>{/each}
		</select>
		<Button onclick={() => (showForm ? (showForm = false) : openForm())}>
			{showForm ? 'Cancelar' : 'Nova tarefa'}
		</Button>
	</div>
</header>

{#if showForm}
	<form class="panel form" onsubmit={create}>
		<div class="grid">
			<Input label="Título" bind:value={titulo} required />
			<Select label="Prioridade" bind:value={prioridade} options={PRIORITY_OPTIONS} />
			<Input label="Prazo" type="date" bind:value={prazo} />
			<Select label="Projeto" bind:value={formProject} options={projectOptions} />
			<Input label="Tags (vírgula)" bind:value={tagsText} placeholder="frontend, urgente" />
		</div>
		<Input label="Descrição" bind:value={descricao} />
		{#if formError}<p class="error">{formError}</p>{/if}
		<div class="actions">
			<Button type="submit" disabled={saving}>{saving ? 'Salvando…' : 'Criar tarefa'}</Button>
		</div>
	</form>
{/if}

{#if loading}
	<p class="muted">Carregando…</p>
{:else if error}
	<p class="error">{error}</p>
{:else}
	<div class="board">
		{#each COLUMNS as col (col)}
			<section class="column">
				<header class="col-head">
					<StatusBadge status={STATUS[col].variant} label={STATUS[col].label} />
					<span class="count">{byStatus(col).length}</span>
				</header>
				<div class="cards">
					{#each byStatus(col) as t (t.id)}
						<article class="card" class:overdue={isOverdue(t.prazo, t.status)}>
							<div class="card-top">
								<StatusBadge
									status={PRIORITY[t.prioridade]?.variant ?? 'info'}
									label={PRIORITY[t.prioridade]?.label ?? t.prioridade}
								/>
								<button class="del" onclick={() => remove(t)} aria-label="Excluir tarefa">✕</button>
							</div>
							<p class="title">{t.titulo}</p>
							{#if t.descricao}<p class="card-desc">{t.descricao}</p>{/if}
							<div class="meta">
								{#if t.project_id}
									<a class="proj" href={`/projetos/${t.project_id}`}>{projectName(t.project_id)}</a>
								{/if}
								{#if t.prazo}
									<span class="prazo" class:late={isOverdue(t.prazo, t.status)}>{prazoLabel(t.prazo)}</span>
								{/if}
							</div>
							{#if t.tags.length}
								<div class="tags">
									{#each t.tags as tag (tag)}<span class="tag">{tag}</span>{/each}
								</div>
							{/if}
							<div class="moves">
								<button
									onclick={() => move(t, -1)}
									disabled={col === COLUMNS[0]}
									aria-label="Mover para a coluna anterior">←</button
								>
								<button
									onclick={() => move(t, 1)}
									disabled={col === COLUMNS[COLUMNS.length - 1]}
									aria-label="Mover para a próxima coluna">→</button
								>
							</div>
						</article>
					{/each}
					{#if byStatus(col).length === 0}
						<p class="col-empty">Vazia</p>
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
	}
	.head-actions {
		display: flex;
		align-items: center;
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
	.filter {
		font-family: var(--font-sans);
		font-size: 13px;
		padding: 8px 10px;
		background-color: var(--color-surface);
		color: var(--color-text);
		border: var(--border-width-1) solid var(--color-border);
		border-radius: var(--radius-md);
		cursor: pointer;
	}
	.filter:hover {
		border-color: var(--color-border-strong);
	}
	.filter:focus {
		outline: none;
		border-color: var(--color-primary);
		box-shadow: var(--shadow-focus);
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
		text-align: center;
		color: var(--color-text-disabled);
		font-size: var(--text-sm);
		padding: var(--space-4) 0;
		margin: 0;
	}
	.card {
		background-color: var(--color-surface);
		border: var(--border-width-1) solid var(--color-border);
		border-radius: var(--radius-md);
		box-shadow: var(--shadow-sm);
		padding: var(--space-3);
		display: flex;
		flex-direction: column;
		gap: var(--space-2);
	}
	.card.overdue {
		border-color: var(--color-danger);
	}
	.card-top {
		display: flex;
		align-items: center;
		justify-content: space-between;
	}
	.del {
		background: none;
		border: none;
		color: var(--color-text-muted);
		font-size: 13px;
		line-height: 1;
		cursor: pointer;
		padding: 2px;
	}
	.del:hover {
		color: var(--color-danger-text);
	}
	.title {
		font-size: 13.5px;
		font-weight: var(--weight-semibold);
		color: var(--color-text);
		margin: 0;
	}
	.card-desc {
		font-size: 12.5px;
		color: var(--color-text-secondary);
		margin: 0;
	}
	.meta {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: var(--space-2) var(--space-3);
		font-size: 12px;
	}
	.proj {
		color: var(--color-link);
		text-decoration: none;
	}
	.prazo {
		font-family: var(--font-mono);
		font-size: 11px;
		color: var(--color-text-muted);
	}
	.prazo.late {
		color: var(--color-danger-text);
		font-weight: var(--weight-semibold);
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
	.moves {
		display: flex;
		justify-content: flex-end;
		gap: var(--space-2);
		border-top: var(--border-width-1) solid var(--color-divider);
		padding-top: var(--space-2);
	}
	.moves button {
		background-color: var(--color-surface);
		border: var(--border-width-1) solid var(--color-border);
		border-radius: var(--radius-sm);
		color: var(--color-text-secondary);
		font-size: 13px;
		line-height: 1;
		padding: 3px 9px;
		cursor: pointer;
		transition: border-color var(--dur-fast) var(--ease-out);
	}
	.moves button:hover:not(:disabled) {
		border-color: var(--color-border-strong);
		color: var(--color-text);
	}
	.moves button:disabled {
		opacity: 0.35;
		cursor: not-allowed;
	}

	@media (max-width: 720px) {
		.board {
			grid-template-columns: 1fr;
		}
	}
</style>
