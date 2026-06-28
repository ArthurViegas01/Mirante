<script>
	import { onMount } from 'svelte';
	import Button from '$lib/components/Button.svelte';
	import Input from '$lib/components/Input.svelte';
	import Textarea from '$lib/components/Textarea.svelte';
	import Skeleton from '$lib/components/Skeleton.svelte';
	import { api } from '$lib/api.js';
	import { toasts } from '$lib/stores/toast.svelte.js';

	// "Meu Perfil" owns the identity of the master CV: name, current/target
	// profession, contact, summary and skills (PUT /api/profile preserves the
	// experiences/education edited under /cv). The Currículo card below pulls a
	// read-only summary from the same record and links to the full editor.
	let nome = $state('');
	let titulo = $state('');
	let tituloAlvo = $state('');
	let contato = $state('');
	let resumo = $state('');
	let skillsText = $state('');
	let savedSkills = $state([]);
	let experiences = $state([]);
	let education = $state([]);

	let loading = $state(true);
	let saving = $state(false);
	let error = $state('');
	let saved = $state(false);

	async function load() {
		loading = true;
		error = '';
		try {
			const p = await api('/api/profile');
			nome = p.nome ?? '';
			titulo = p.titulo ?? '';
			tituloAlvo = p.titulo_alvo ?? '';
			contato = p.contato ?? '';
			resumo = p.resumo ?? '';
			savedSkills = p.skills ?? [];
			skillsText = savedSkills.join(', ');
			experiences = p.experiences ?? [];
			education = p.education ?? [];
		} catch (e) {
			error = e.message;
			toasts.error(e.message);
		} finally {
			loading = false;
		}
	}

	onMount(load);

	async function save(e) {
		e.preventDefault();
		saving = true;
		error = '';
		saved = false;
		try {
			const skills = skillsText
				.split(',')
				.map((s) => s.trim())
				.filter(Boolean);
			const p = await api('/api/profile', {
				method: 'PUT',
				body: { nome, titulo, titulo_alvo: tituloAlvo, contato, resumo, skills }
			});
			savedSkills = p.skills ?? [];
			skillsText = savedSkills.join(', ');
			saved = true;
			toasts.success('Perfil salvo');
		} catch (e) {
			error = e.message;
			toasts.error(e.message);
		} finally {
			saving = false;
		}
	}
</script>

<header class="page-head">
	<h1>Meu Perfil</h1>
</header>

{#if loading}
	<div aria-hidden="true">
		<section class="panel sk-panel">
			<Skeleton w="160px" h="18px" />
			<div class="grid">
				{#each Array(3) as _, j (j)}<Skeleton w="100%" h="34px" radius="var(--radius-md)" block />{/each}
			</div>
			<Skeleton w="100%" h="64px" radius="var(--radius-md)" block />
		</section>
	</div>
{:else}
	<form onsubmit={save}>
		<section class="panel">
			<h2>Identidade</h2>
			<div class="grid">
				<Input label="Nome" bind:value={nome} />
				<Input label="Profissão atual" bind:value={titulo} placeholder="Dev Backend Pleno" />
				<Input label="Profissão almejada" bind:value={tituloAlvo} placeholder="Staff Engineer" />
			</div>
			<Input
				label="Contato (e-mail · telefone · localização · links)"
				bind:value={contato}
				placeholder="arthur@email.com · +55 51 … · Porto Alegre · github.com/…"
			/>
			<Textarea
				label="Resumo profissional"
				bind:value={resumo}
				rows={4}
				placeholder="Um parágrafo sobre você, sua stack e o que busca."
			/>
			<div class="field">
				<span class="label">Skills (vírgula): base do <strong>% de aderência</strong> das vagas</span>
				<Input bind:value={skillsText} placeholder="Go, React, PostgreSQL, Docker…" />
			</div>
			{#if savedSkills.length}
				<div class="skills">
					{#each savedSkills as s (s)}<span class="skill">{s}</span>{/each}
				</div>
			{/if}
		</section>

		{#if error}<p class="error">{error}</p>{/if}
		<div class="actions">
			{#if saved}<span class="ok">Salvo ✓</span>{/if}
			<Button type="submit" disabled={saving}>{saving ? 'Salvando…' : 'Salvar perfil'}</Button>
		</div>
	</form>

	<section class="panel cv-card">
		<div class="panel-head">
			<h2>Currículo</h2>
			<a class="cv-link" href="/cv">Editar currículo completo →</a>
		</div>
		<p class="muted">
			Experiências, formação e exportação (PDF/DOCX) ficam no currículo. Ele usa a identidade acima.
		</p>
		<div class="cv-stats">
			<div class="stat"><span class="num">{experiences.length}</span> experiências</div>
			<div class="stat"><span class="num">{education.length}</span> formações</div>
		</div>
		{#if experiences.length}
			<ul class="exp-list">
				{#each experiences.slice(0, 3) as e, i (i)}
					<li>{e.cargo || 'Cargo'}{e.empresa ? ` · ${e.empresa}` : ''}</li>
				{/each}
				{#if experiences.length > 3}<li class="more">+ {experiences.length - 3} no currículo</li>{/if}
			</ul>
		{/if}
	</section>
{/if}

<style>
	.page-head {
		margin-bottom: var(--space-6);
	}
	h1 {
		font-size: var(--text-2xl);
		font-weight: var(--weight-medium);
		letter-spacing: var(--tracking-snug);
		color: var(--color-text);
		margin: 0;
	}
	h2 {
		font-size: var(--text-lg);
		font-weight: var(--weight-medium);
		color: var(--color-text);
		margin: 0 0 var(--space-4);
	}
	.muted {
		color: var(--color-text-secondary);
		font-size: var(--text-sm);
		margin: 0;
	}
	.panel {
		background-color: var(--color-surface);
		border: var(--border-width-1) solid var(--color-border);
		border-radius: var(--radius-lg);
		box-shadow: var(--shadow-sm);
		padding: var(--space-6);
		margin-bottom: var(--space-5);
		display: flex;
		flex-direction: column;
		gap: var(--space-4);
		max-width: var(--max-prose);
	}
	.panel-head {
		display: flex;
		align-items: baseline;
		justify-content: space-between;
		gap: var(--space-4);
	}
	.panel-head h2 {
		margin: 0;
	}
	.grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
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
	.sk-panel {
		gap: var(--space-3);
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
	.actions {
		display: flex;
		align-items: center;
		justify-content: flex-end;
		gap: var(--space-3);
		max-width: var(--max-prose);
		margin-bottom: var(--space-6);
	}
	.ok {
		font-size: var(--text-sm);
		color: var(--color-success-text);
	}
	.error {
		color: var(--color-danger-text);
		font-size: var(--text-sm);
		max-width: var(--max-prose);
	}
	.cv-link {
		font-size: var(--text-sm);
		color: var(--color-link);
		text-decoration: none;
		white-space: nowrap;
	}
	.cv-link:hover {
		text-decoration: underline;
	}
	.cv-stats {
		display: flex;
		gap: var(--space-6);
	}
	.stat {
		font-size: var(--text-sm);
		color: var(--color-text-secondary);
	}
	.stat .num {
		font-size: var(--text-lg);
		font-weight: var(--weight-semibold);
		color: var(--color-text);
		margin-right: 4px;
	}
	.exp-list {
		margin: 0;
		padding-left: var(--space-4);
		font-size: var(--text-sm);
		color: var(--color-text-secondary);
		display: flex;
		flex-direction: column;
		gap: 2px;
	}
	.exp-list .more {
		list-style: none;
		margin-left: calc(-1 * var(--space-4));
		color: var(--color-text-muted);
	}
</style>
