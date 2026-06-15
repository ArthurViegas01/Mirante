<script>
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import Button from '$lib/components/Button.svelte';
	import Input from '$lib/components/Input.svelte';
	import Select from '$lib/components/Select.svelte';
	import StatusBadge from '$lib/components/StatusBadge.svelte';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import Skeleton from '$lib/components/Skeleton.svelte';
	import ProjectStacks from '$lib/components/ProjectStacks.svelte';
	import ProjectCosts from '$lib/components/ProjectCosts.svelte';
	import { api } from '$lib/api.js';
	import { toasts } from '$lib/stores/toast.svelte.js';
	import { confirm } from '$lib/stores/confirm.svelte.js';
	import { STATUS, STATUS_OPTIONS, LINK_KINDS } from '$lib/projectStatus.js';
	import { STATUS as TASK_STATUS, isOverdue, prazoLabel } from '$lib/taskMeta.js';

	let id = $derived($page.params.id);

	let project = $state(null);
	let tasks = $state([]);
	let loading = $state(true);
	let error = $state('');
	let status = $state('');

	let openTasks = $derived(tasks.filter((t) => t.status !== 'feito'));

	let linkLabel = $state('');
	let linkUrl = $state('');
	let linkKind = $state('other');
	let linkError = $state('');

	async function load(pid) {
		loading = true;
		error = '';
		try {
			const [p, tr] = await Promise.all([
				api(`/api/projects/${pid}`),
				api(`/api/tasks?project=${pid}`)
			]);
			project = p;
			status = project.status;
			tasks = tr.tasks;
		} catch (e) {
			error = e.message;
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		load(id);
	});

	async function saveStatus() {
		try {
			await api(`/api/projects/${id}`, { method: 'PATCH', body: { status } });
			await load(id);
			toasts.success('Status atualizado');
		} catch (e) {
			error = e.message;
			toasts.error(e.message);
		}
	}

	async function addLink(e) {
		e.preventDefault();
		linkError = '';
		try {
			await api(`/api/projects/${id}/links`, {
				method: 'POST',
				body: { label: linkLabel, url: linkUrl, kind: linkKind }
			});
			linkLabel = linkUrl = '';
			linkKind = 'other';
			await load(id);
			toasts.success('Link adicionado');
		} catch (e) {
			linkError = e.message;
			toasts.error(e.message);
		}
	}

	async function removeLink(linkId) {
		const ok = await confirm.ask({
			title: 'Remover link?',
			message: 'O link será removido deste projeto.',
			confirmLabel: 'Remover',
			danger: true
		});
		if (!ok) return;
		try {
			await api(`/api/projects/${id}/links/${linkId}`, { method: 'DELETE' });
			await load(id);
			toasts.success('Link removido');
		} catch (e) {
			toasts.error(e.message);
		}
	}

	async function remove() {
		const ok = await confirm.ask({
			title: 'Excluir projeto?',
			message:
				'Esta ação não pode ser desfeita. Os serviços do monitor vão junto; as tarefas são desvinculadas (não apagadas).',
			confirmLabel: 'Excluir',
			danger: true
		});
		if (!ok) return;
		try {
			await api(`/api/projects/${id}`, { method: 'DELETE' });
			toasts.success('Projeto excluído');
			goto('/projetos');
		} catch (e) {
			error = e.message;
			toasts.error(e.message);
		}
	}
</script>

<a class="back" href="/projetos">← Projetos</a>

{#if loading}
	<div class="sk-detail" aria-hidden="true">
		<div class="sk-head">
			<Skeleton w="240px" h="28px" />
			<Skeleton w="84px" h="22px" radius="var(--radius-full)" />
		</div>
		<Skeleton w="70%" h="14px" block />
		<div class="sk-meta">
			<Skeleton w="180px" h="13px" />
			<Skeleton w="120px" h="13px" />
		</div>
		<div class="panel sk-panel">
			<Skeleton w="120px" h="18px" />
			<Skeleton w="100%" h="40px" radius="var(--radius-md)" block />
		</div>
		<div class="panel sk-panel">
			<Skeleton w="90px" h="18px" />
			<Skeleton w="100%" h="40px" radius="var(--radius-md)" block />
		</div>
	</div>
{:else if error}
	<p class="error">{error}</p>
{:else if project}
	<header class="head">
		<div>
			<h1>{project.nome}</h1>
			{#if project.codinome}<span class="codinome">{project.codinome}</span>{/if}
		</div>
		<StatusBadge
			status={STATUS[project.status]?.variant ?? 'info'}
			label={STATUS[project.status]?.label ?? project.status}
		/>
	</header>

	{#if project.descricao}<p class="desc">{project.descricao}</p>{/if}

	<div class="meta-row">
		{#if project.repo}<a class="repo" href={project.repo} target="_blank" rel="noreferrer">{project.repo} ↗</a>{/if}
		<span class="vis">Visibilidade: {project.visibilidade}</span>
	</div>

	{#if project.tags.length}
		<div class="tags">
			{#each project.tags as t (t)}<span class="tag">{t}</span>{/each}
		</div>
	{/if}

	<section class="panel">
		<h2>Status</h2>
		<div class="status-row">
			<Select bind:value={status} options={STATUS_OPTIONS} />
			<Button variant="secondary" onclick={saveStatus} disabled={status === project.status}>Salvar</Button>
		</div>
	</section>

	<section class="panel">
		<h2>Links</h2>
		{#if project.links.length}
			<ul class="links">
				{#each project.links as l (l.id)}
					<li>
						<a href={l.url} target="_blank" rel="noreferrer">{l.label} ↗</a>
						<span class="kind">{l.kind}</span>
						<button class="del" onclick={() => removeLink(l.id)} aria-label="Remover link">remover</button>
					</li>
				{/each}
			</ul>
		{:else}
			<EmptyState
				compact
				title="Nenhum link ainda"
				description="Adicione repositórios, deploys ou docs relevantes abaixo."
			/>
		{/if}
		<form class="link-form" onsubmit={addLink}>
			<Input label="Label" bind:value={linkLabel} required />
			<Input label="URL" type="url" bind:value={linkUrl} placeholder="https://…" required />
			<Select label="Tipo" bind:value={linkKind} options={LINK_KINDS} />
			<div class="link-submit"><Button type="submit" size="sm">Adicionar</Button></div>
		</form>
		{#if linkError}<p class="error">{linkError}</p>{/if}
	</section>

	<ProjectStacks {id} />

	<ProjectCosts {id} />

	<section class="panel">
		<div class="panel-head">
			<h2>Tarefas abertas</h2>
			<a class="board-link" href={`/tarefas?project=${id}`}>Ver no quadro →</a>
		</div>
		{#if openTasks.length}
			<ul class="tasks">
				{#each openTasks as t (t.id)}
					<li>
						<StatusBadge
							status={TASK_STATUS[t.status]?.variant ?? 'info'}
							label={TASK_STATUS[t.status]?.label ?? t.status}
						/>
						<span class="task-title">{t.titulo}</span>
						{#if t.prazo}
							<span class="task-prazo" class:late={isOverdue(t.prazo, t.status)}>{prazoLabel(t.prazo)}</span>
						{/if}
					</li>
				{/each}
			</ul>
		{:else}
			<EmptyState compact title="Nenhuma tarefa aberta" description="Tudo em dia por aqui.">
				{#snippet children()}
					<a class="board-link" href={`/tarefas?project=${id}`}>Criar no quadro →</a>
				{/snippet}
			</EmptyState>
		{/if}
	</section>

	<section class="danger-zone">
		<Button variant="ghost" onclick={remove}>Excluir projeto</Button>
	</section>
{/if}

<style>
	.back {
		font-size: var(--text-sm);
		color: var(--color-text-muted);
		text-decoration: none;
		display: inline-block;
		margin-bottom: var(--space-4);
	}
	.back:hover {
		color: var(--color-text);
	}
	.head {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: var(--space-4);
		margin-bottom: var(--space-3);
	}
	h1 {
		font-size: var(--text-2xl);
		font-weight: var(--weight-medium);
		letter-spacing: var(--tracking-snug);
		color: var(--color-text);
		margin: 0;
	}
	.codinome {
		font-family: var(--font-mono);
		font-size: var(--text-sm);
		color: var(--color-text-muted);
	}
	.desc {
		color: var(--color-text-secondary);
		max-width: var(--max-prose);
		margin: 0 0 var(--space-4);
	}
	.meta-row {
		display: flex;
		flex-wrap: wrap;
		gap: var(--space-4);
		font-size: var(--text-sm);
		margin-bottom: var(--space-4);
	}
	.repo {
		color: var(--color-link);
		text-decoration: none;
	}
	.vis {
		color: var(--color-text-muted);
		text-transform: capitalize;
	}
	.tags {
		display: flex;
		flex-wrap: wrap;
		gap: 4px;
		margin-bottom: var(--space-6);
	}
	.tag {
		font-family: var(--font-mono);
		font-size: 11px;
		padding: 2px 7px;
		border-radius: var(--radius-full);
		background-color: var(--color-divider);
		color: var(--color-text-secondary);
	}
	.panel {
		background-color: var(--color-surface);
		border: var(--border-width-1) solid var(--color-border);
		border-radius: var(--radius-lg);
		box-shadow: var(--shadow-sm);
		padding: var(--space-6);
		margin-bottom: var(--space-5);
	}
	h2 {
		font-size: var(--text-lg);
		font-weight: var(--weight-medium);
		color: var(--color-text);
		margin: 0 0 var(--space-4);
	}
	.status-row {
		display: flex;
		align-items: flex-end;
		gap: var(--space-3);
	}
	.links {
		list-style: none;
		margin: 0 0 var(--space-4);
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: var(--space-2);
	}
	.links li {
		display: flex;
		align-items: center;
		gap: var(--space-3);
		font-size: var(--text-sm);
	}
	.links a {
		color: var(--color-link);
		text-decoration: none;
	}
	.kind {
		font-family: var(--font-mono);
		font-size: 11px;
		color: var(--color-text-muted);
	}
	.del {
		margin-left: auto;
		background: none;
		border: none;
		color: var(--color-danger-text);
		font-size: var(--text-sm);
		cursor: pointer;
	}
	.link-form {
		display: grid;
		grid-template-columns: 1fr 1fr 1fr auto;
		gap: var(--space-3);
		align-items: end;
		border-top: var(--border-width-1) solid var(--color-divider);
		padding-top: var(--space-4);
	}
	.link-submit {
		display: flex;
	}
	.sk-detail {
		display: flex;
		flex-direction: column;
		gap: var(--space-4);
	}
	.sk-head {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: var(--space-4);
	}
	.sk-meta {
		display: flex;
		gap: var(--space-4);
	}
	.sk-panel {
		display: flex;
		flex-direction: column;
		gap: var(--space-4);
		margin-bottom: 0;
	}
	.error {
		color: var(--color-danger-text);
		font-size: var(--text-sm);
	}
	.danger-zone {
		margin-top: var(--space-8);
	}
	.panel-head {
		display: flex;
		align-items: baseline;
		justify-content: space-between;
		gap: var(--space-4);
		margin-bottom: var(--space-4);
	}
	.panel-head h2 {
		margin: 0;
	}
	.board-link {
		font-size: var(--text-sm);
		color: var(--color-link);
		text-decoration: none;
		white-space: nowrap;
	}
	.tasks {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: var(--space-3);
	}
	.tasks li {
		display: flex;
		align-items: center;
		gap: var(--space-3);
		font-size: var(--text-sm);
	}
	.task-title {
		color: var(--color-text);
	}
	.task-prazo {
		margin-left: auto;
		font-family: var(--font-mono);
		font-size: 11px;
		color: var(--color-text-muted);
		white-space: nowrap;
	}
	.task-prazo.late {
		color: var(--color-danger-text);
		font-weight: var(--weight-semibold);
	}
	@media (max-width: 640px) {
		.link-form {
			grid-template-columns: 1fr;
		}
	}
</style>
