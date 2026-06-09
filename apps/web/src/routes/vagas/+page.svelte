<script>
	import { onMount } from 'svelte';
	import Button from '$lib/components/Button.svelte';
	import Input from '$lib/components/Input.svelte';
	import Select from '$lib/components/Select.svelte';
	import ProfileHeadline from '$lib/components/ProfileHeadline.svelte';
	import { api } from '$lib/api.js';
	import { MODELO_OPTIONS, modeloLabel } from '$lib/jobMeta.js';

	let jobs = $state([]);
	let loading = $state(true);
	let error = $state('');
	let showForm = $state(false);

	let titulo = $state('');
	let empresa = $state('');
	let descricao = $state('');
	let url = $state('');
	let localizacao = $state('');
	let modelo = $state('indefinido');
	let senioridade = $state('');
	let saving = $state(false);
	let formError = $state('');

	let importing = $state(false);
	let importError = $state('');
	let importFonte = $state('');

	let enrichingId = $state('');

	async function load() {
		loading = true;
		error = '';
		try {
			const res = await api('/api/jobs');
			jobs = res.jobs;
		} catch (e) {
			error = e.message;
		} finally {
			loading = false;
		}
	}

	onMount(load);

	function resetForm() {
		titulo = empresa = descricao = url = localizacao = senioridade = '';
		modelo = 'indefinido';
		formError = importError = importFonte = '';
	}

	async function importLink() {
		if (!url) return;
		importing = true;
		importError = '';
		importFonte = '';
		try {
			const d = await api('/api/jobs/import', { method: 'POST', body: { url } });
			titulo = d.titulo || titulo;
			empresa = d.empresa || empresa;
			descricao = d.descricao || descricao;
			localizacao = d.localizacao || localizacao;
			senioridade = d.senioridade || senioridade;
			if (d.modelo && d.modelo !== 'indefinido') modelo = d.modelo;
			importFonte = d.fonte;
		} catch (e) {
			importError = e.message;
		} finally {
			importing = false;
		}
	}

	async function create(e) {
		e.preventDefault();
		saving = true;
		formError = '';
		try {
			await api('/api/jobs', {
				method: 'POST',
				body: { titulo, empresa, descricao, url, localizacao, modelo, senioridade }
			});
			showForm = false;
			resetForm();
			await load();
		} catch (e) {
			formError = e.message;
		} finally {
			saving = false;
		}
	}

	async function enrich(job) {
		enrichingId = job.id;
		error = '';
		try {
			await api(`/api/jobs/${job.id}/enrich`, { method: 'POST' });
			await load();
		} catch (e) {
			error = e.message;
		} finally {
			enrichingId = '';
		}
	}

	async function remove(job) {
		if (!confirm(`Excluir a vaga "${job.titulo}"?`)) return;
		try {
			await api(`/api/jobs/${job.id}`, { method: 'DELETE' });
			await load();
		} catch (e) {
			error = e.message;
		}
	}
</script>

<header class="page-head">
	<div>
		<p class="eyebrow">Vagas</p>
		<h1>Vagas</h1>
		<ProfileHeadline />
	</div>
	<Button onclick={() => (showForm ? (showForm = false) : (showForm = true))}>
		{showForm ? 'Cancelar' : 'Nova vaga'}
	</Button>
</header>

{#if showForm}
	<form class="panel form" onsubmit={create}>
		<div class="import-row">
			<Input
				label="Link da vaga (LinkedIn, etc.) — preenche o formulário automaticamente"
				type="url"
				bind:value={url}
				placeholder="https://www.linkedin.com/jobs/view/…"
			/>
			<div class="import-btn">
				<Button variant="secondary" onclick={importLink} disabled={importing || !url}>
					{importing ? 'Buscando…' : 'Buscar do link'}
				</Button>
			</div>
		</div>
		{#if importError}<p class="error">{importError}</p>{/if}
		{#if importFonte}<p class="hint">Preenchido a partir do link (via {importFonte}). Confira e ajuste antes de salvar.</p>{/if}

		<div class="grid">
			<Input label="Título / cargo" bind:value={titulo} required />
			<Input label="Empresa" bind:value={empresa} />
			<Select label="Modelo" bind:value={modelo} options={MODELO_OPTIONS} />
			<Input label="Senioridade" bind:value={senioridade} placeholder="júnior, pleno…" />
			<Input label="Localização" bind:value={localizacao} />
		</div>
		<label class="field">
			<span class="label">Descrição (cole o texto da vaga — as skills são extraídas daqui)</span>
			<textarea bind:value={descricao} rows="6" placeholder="Responsabilidades, requisitos, stack…"></textarea>
		</label>
		{#if formError}<p class="error">{formError}</p>{/if}
		<div class="actions">
			<Button type="submit" disabled={saving}>{saving ? 'Salvando…' : 'Adicionar vaga'}</Button>
		</div>
	</form>
{/if}

{#if error}<p class="error">{error}</p>{/if}

{#if loading}
	<p class="muted">Carregando…</p>
{:else if jobs.length === 0}
	<div class="panel empty">Nenhuma vaga ainda. Adicione a primeira e cole a descrição.</div>
{:else}
	<div class="list">
		{#each jobs as job (job.id)}
			<article class="panel card">
				<div class="card-head">
					<div>
						<h2>{job.titulo}</h2>
						{#if job.empresa}<span class="empresa">{job.empresa}</span>{/if}
					</div>
					<span class="chip">{modeloLabel(job.modelo)}</span>
				</div>

				<div class="meta">
					{#if job.senioridade}<span>{job.senioridade}</span>{/if}
					{#if job.localizacao}<span>{job.localizacao}</span>{/if}
					{#if job.url}<a href={job.url} target="_blank" rel="noreferrer">link ↗</a>{/if}
				</div>

				{#if job.resumo}<p class="resumo">{job.resumo}</p>{/if}

				{#if job.skills.length}
					<div class="skills">
						{#each job.skills as s (s)}<span class="skill">{s}</span>{/each}
					</div>
				{:else}
					<p class="muted small">Nenhuma skill reconhecida na descrição.</p>
				{/if}

				<div class="card-actions">
					<Button size="sm" variant="secondary" onclick={() => enrich(job)} disabled={enrichingId === job.id}>
						{enrichingId === job.id ? 'Enriquecendo…' : '✨ Enriquecer com IA'}
					</Button>
					<button class="del" onclick={() => remove(job)}>Excluir</button>
				</div>
			</article>
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
	.muted {
		color: var(--color-text-secondary);
	}
	.muted.small {
		font-size: var(--text-sm);
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
		padding: 10px;
		background-color: var(--color-surface);
		color: var(--color-text);
		border: var(--border-width-1) solid var(--color-border);
		border-radius: var(--radius-md);
		resize: vertical;
		min-height: 96px;
	}
	textarea:focus {
		outline: none;
		border-color: var(--color-primary);
		box-shadow: var(--shadow-focus);
	}
	.actions {
		display: flex;
		justify-content: flex-end;
	}
	.import-row {
		display: grid;
		grid-template-columns: 1fr auto;
		gap: var(--space-3);
		align-items: end;
	}
	.import-btn {
		display: flex;
	}
	.hint {
		margin: 0;
		font-size: var(--text-sm);
		color: var(--color-accent);
	}
	.empty {
		padding: var(--space-8);
		text-align: center;
		color: var(--color-text-muted);
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
	.card-head {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: var(--space-4);
	}
	h2 {
		font-size: var(--text-lg);
		font-weight: var(--weight-semibold);
		color: var(--color-text);
		margin: 0;
	}
	.empresa {
		font-size: var(--text-sm);
		color: var(--color-text-secondary);
	}
	.chip {
		font-family: var(--font-mono);
		font-size: 11px;
		padding: 3px 9px;
		border-radius: var(--radius-full);
		background-color: var(--color-divider);
		color: var(--color-text-secondary);
		white-space: nowrap;
	}
	.meta {
		display: flex;
		flex-wrap: wrap;
		gap: var(--space-3);
		font-size: var(--text-sm);
		color: var(--color-text-muted);
	}
	.meta a {
		color: var(--color-link);
		text-decoration: none;
	}
	.resumo {
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
	.card-actions {
		display: flex;
		align-items: center;
		gap: var(--space-3);
		border-top: var(--border-width-1) solid var(--color-divider);
		padding-top: var(--space-3);
	}
	.del {
		margin-left: auto;
		background: none;
		border: none;
		color: var(--color-text-muted);
		font-size: var(--text-sm);
		cursor: pointer;
	}
	.del:hover {
		color: var(--color-danger-text);
	}
</style>
