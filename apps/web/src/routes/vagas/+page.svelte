<script>
	import { onMount } from 'svelte';
	import Button from '$lib/components/Button.svelte';
	import Input from '$lib/components/Input.svelte';
	import Textarea from '$lib/components/Textarea.svelte';
	import Select from '$lib/components/Select.svelte';
	import StatusBadge from '$lib/components/StatusBadge.svelte';
	import ProfileHeadline from '$lib/components/ProfileHeadline.svelte';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import Skeleton from '$lib/components/Skeleton.svelte';
	import { api } from '$lib/api.js';
	import { toasts } from '$lib/stores/toast.svelte.js';
	import { confirm } from '$lib/stores/confirm.svelte.js';
	import { MODELO_OPTIONS, modeloLabel } from '$lib/jobMeta.js';
	import { aderencia, aderenciaVariant } from '$lib/aderencia.js';

	let jobs = $state([]);
	let mySkills = $state([]);
	let loading = $state(true);
	let error = $state('');
	let showForm = $state(false);

	let mySet = $derived(new Set(mySkills.map((s) => s.toLowerCase())));

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
	let adaptingId = $state('');
	let adaptId = $state('');
	let adaptResult = $state(null);
	let trackingId = $state('');
	let trackedIds = $state(new Set());

	async function load() {
		loading = true;
		error = '';
		try {
			const [jres, profile] = await Promise.all([api('/api/jobs'), api('/api/profile')]);
			jobs = jres.jobs;
			mySkills = profile.skills ?? [];
		} catch (e) {
			error = e.message;
			toasts.error(e.message);
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
			toasts.success('Dados importados do link');
		} catch (e) {
			importError = e.message;
			toasts.error(e.message);
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
			toasts.success('Vaga adicionada');
			await load();
		} catch (e) {
			formError = e.message;
			toasts.error(e.message);
		} finally {
			saving = false;
		}
	}

	async function enrich(job) {
		enrichingId = job.id;
		error = '';
		try {
			await api(`/api/jobs/${job.id}/enrich`, { method: 'POST' });
			toasts.success('Vaga enriquecida');
			await load();
		} catch (e) {
			error = e.message;
			toasts.error(e.message);
		} finally {
			enrichingId = '';
		}
	}

	async function adaptCV(job) {
		adaptingId = job.id;
		error = '';
		try {
			adaptResult = await api('/api/cv/adapt', {
				method: 'POST',
				body: {
					titulo: job.titulo,
					empresa: job.empresa,
					descricao: job.descricao,
					skills: job.skills
				}
			});
			adaptId = job.id;
			toasts.success('CV adaptado para esta vaga');
		} catch (e) {
			error = e.message;
			toasts.error(e.message);
		} finally {
			adaptingId = '';
		}
	}

	async function trackJob(job) {
		trackingId = job.id;
		error = '';
		try {
			await api('/api/applications', {
				method: 'POST',
				body: { job_id: job.id, titulo: job.titulo, empresa: job.empresa }
			});
			trackedIds = new Set([...trackedIds, job.id]);
			toasts.success('Vaga adicionada ao pipeline');
		} catch (e) {
			error = e.message;
			toasts.error(e.message);
		} finally {
			trackingId = '';
		}
	}

	async function remove(job) {
		if (
			!(await confirm.ask({
				title: 'Excluir vaga?',
				message: `A vaga "${job.titulo}" será removida em definitivo.`,
				confirmLabel: 'Excluir',
				danger: true
			}))
		)
			return;
		try {
			await api(`/api/jobs/${job.id}`, { method: 'DELETE' });
			toasts.success('Vaga excluída');
			await load();
		} catch (e) {
			error = e.message;
			toasts.error(e.message);
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
				label="Link da vaga (LinkedIn, etc.) preenche o formulário automaticamente"
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
		<Textarea
			label="Descrição (cole o texto da vaga, as skills são extraídas daqui)"
			bind:value={descricao}
			rows={6}
			placeholder="Responsabilidades, requisitos, stack…"
		/>
		{#if formError}<p class="error">{formError}</p>{/if}
		<div class="actions">
			<Button type="submit" disabled={saving}>{saving ? 'Salvando…' : 'Adicionar vaga'}</Button>
		</div>
	</form>
{/if}

{#if error}<p class="error">{error}</p>{/if}

{#if loading}
	<div class="list" aria-hidden="true">
		{#each Array(3) as _, i (i)}
			<article class="panel card sk-card">
				<div class="card-head">
					<Skeleton w="45%" h="18px" />
					<Skeleton w="80px" h="20px" radius="var(--radius-full)" />
				</div>
				<Skeleton w="30%" h="12px" />
				<Skeleton w="90%" h="12px" block />
				<div class="skills">
					{#each Array(4) as __, j (j)}<Skeleton w="56px" h="18px" radius="var(--radius-full)" />{/each}
				</div>
			</article>
		{/each}
	</div>
{:else if jobs.length === 0}
	<EmptyState
		title="Nenhuma vaga ainda"
		description="Adicione a primeira vaga: cole o link para preencher automaticamente ou registre a descrição manualmente."
	>
		{#snippet children()}
			<Button onclick={() => (showForm = true)}>Nova vaga</Button>
		{/snippet}
	</EmptyState>
{:else}
	<div class="list">
		{#each jobs as job (job.id)}
			{@const a = aderencia(job.skills, mySkills)}
				<article class="panel card">
				<div class="card-head">
					<div>
						<h2>{job.titulo}</h2>
						{#if job.empresa}<span class="empresa">{job.empresa}</span>{/if}
					</div>
					<div class="head-badges">
						{#if mySkills.length && a.score !== null}
							<StatusBadge status={aderenciaVariant(a.score)} label={`${a.score}% match`} />
						{/if}
						<span class="chip">{modeloLabel(job.modelo)}</span>
					</div>
				</div>

				<div class="meta">
					{#if job.senioridade}<span>{job.senioridade}</span>{/if}
					{#if job.localizacao}<span>{job.localizacao}</span>{/if}
					{#if job.url}<a href={job.url} target="_blank" rel="noreferrer">link ↗</a>{/if}
				</div>

				{#if job.resumo}<p class="resumo">{job.resumo}</p>{/if}

				{#if job.skills.length}
					<div class="skills">
						{#each job.skills as s (s)}
							<span
								class="skill"
								class:have={mySet.has(s.toLowerCase())}
								class:lack={mySkills.length && !mySet.has(s.toLowerCase())}>{s}</span
							>
						{/each}
					</div>
				{:else}
					<p class="muted small">Nenhuma skill reconhecida na descrição.</p>
				{/if}

				<div class="card-actions">
					<Button size="sm" variant="secondary" onclick={() => enrich(job)} disabled={enrichingId === job.id}>
						{enrichingId === job.id ? 'Enriquecendo…' : '✨ Enriquecer com IA'}
					</Button>
					<Button size="sm" variant="secondary" onclick={() => adaptCV(job)} disabled={adaptingId === job.id}>
						{adaptingId === job.id ? 'Adaptando…' : '🎯 Adaptar CV'}
					</Button>
					{#if trackedIds.has(job.id)}
						<a class="tracked" href="/candidaturas">✓ No pipeline</a>
					{:else}
						<Button size="sm" variant="secondary" onclick={() => trackJob(job)} disabled={trackingId === job.id}>
							{trackingId === job.id ? '…' : 'Acompanhar'}
						</Button>
					{/if}
					<button class="del" onclick={() => remove(job)}>Excluir</button>
				</div>

				{#if adaptId === job.id && adaptResult}
					<div class="adapt">
						<p class="adapt-label">Resumo adaptado para esta vaga</p>
						<p class="adapt-resumo">{adaptResult.resumo_adaptado}</p>
						<div class="adapt-cols">
							{#if adaptResult.pontos_fortes?.length}
								<div class="adapt-col">
									<p class="adapt-label ok">Pontos fortes</p>
									<ul>{#each adaptResult.pontos_fortes as p (p)}<li>{p}</li>{/each}</ul>
								</div>
							{/if}
							{#if adaptResult.lacunas?.length}
								<div class="adapt-col">
									<p class="adapt-label gap">Lacunas</p>
									<ul>{#each adaptResult.lacunas as l (l)}<li>{l}</li>{/each}</ul>
								</div>
							{/if}
						</div>
						{#if adaptResult.dica}<p class="adapt-dica">💡 {adaptResult.dica}</p>{/if}
					</div>
				{/if}
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
	.head-badges {
		display: flex;
		align-items: center;
		gap: var(--space-2);
		flex-shrink: 0;
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
	.tracked {
		font-size: var(--text-sm);
		color: var(--color-success-text);
		text-decoration: none;
		align-self: center;
	}
	.adapt {
		margin-top: var(--space-3);
		padding: var(--space-4);
		border: var(--border-width-1) solid var(--color-border);
		border-radius: var(--radius-md);
		background-color: var(--color-surface-sunken);
		display: flex;
		flex-direction: column;
		gap: var(--space-3);
	}
	.adapt-label {
		font-family: var(--font-mono);
		font-size: 11px;
		letter-spacing: 0.06em;
		text-transform: uppercase;
		color: var(--color-text-muted);
		margin: 0;
	}
	.adapt-label.ok {
		color: var(--color-success-text);
	}
	.adapt-label.gap {
		color: var(--color-warning-text);
	}
	.adapt-resumo {
		margin: 0;
		font-family: var(--font-serif);
		font-size: var(--text-base);
		color: var(--color-text);
		line-height: var(--leading-relaxed);
	}
	.adapt-cols {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: var(--space-4);
	}
	.adapt-col ul {
		margin: var(--space-1) 0 0;
		padding-left: var(--space-4);
		font-size: var(--text-sm);
		color: var(--color-text-secondary);
		display: flex;
		flex-direction: column;
		gap: 2px;
	}
	.adapt-dica {
		margin: 0;
		font-size: var(--text-sm);
		color: var(--color-text-secondary);
	}
	@media (max-width: 640px) {
		.adapt-cols {
			grid-template-columns: 1fr;
		}
	}
</style>
