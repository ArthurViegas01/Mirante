<script>
	import { onMount } from 'svelte';
	import Button from '$lib/components/Button.svelte';
	import Input from '$lib/components/Input.svelte';
	import { api } from '$lib/api.js';

	let nome = $state('');
	let titulo = $state('');
	let tituloAlvo = $state('');
	let resumo = $state('');
	let skillsText = $state('');
	let savedSkills = $state([]);

	let loading = $state(true);
	let saving = $state(false);
	let error = $state('');
	let savedAt = $state(0);

	async function load() {
		loading = true;
		error = '';
		try {
			const p = await api('/api/profile');
			nome = p.nome ?? '';
			titulo = p.titulo ?? '';
			tituloAlvo = p.titulo_alvo ?? '';
			resumo = p.resumo ?? '';
			savedSkills = p.skills ?? [];
			skillsText = savedSkills.join(', ');
		} catch (e) {
			error = e.message;
		} finally {
			loading = false;
		}
	}

	onMount(load);

	async function save(e) {
		e.preventDefault();
		saving = true;
		error = '';
		try {
			const skills = skillsText
				.split(',')
				.map((s) => s.trim())
				.filter(Boolean);
			const p = await api('/api/profile', {
				method: 'PUT',
				body: { nome, titulo, titulo_alvo: tituloAlvo, resumo, skills }
			});
			savedSkills = p.skills ?? [];
			skillsText = savedSkills.join(', '); // reflect normalized/deduped
			savedAt = Date.now();
		} catch (e) {
			error = e.message;
		} finally {
			saving = false;
		}
	}
</script>

<header class="page-head">
	<div>
		<p class="eyebrow">CV</p>
		<h1>CV mestre</h1>
	</div>
</header>

{#if loading}
	<p class="muted">Carregando…</p>
{:else}
	<form class="panel form" onsubmit={save}>
		<div class="grid">
			<Input label="Nome" bind:value={nome} />
			<Input label="Profissão atual" bind:value={titulo} placeholder="Dev Backend Pleno" />
			<Input label="Profissão almejada" bind:value={tituloAlvo} placeholder="Staff Engineer" />
		</div>

		<label class="field">
			<span class="label">Resumo profissional</span>
			<textarea bind:value={resumo} rows="4" placeholder="Um parágrafo sobre você, sua stack e o que busca."></textarea>
		</label>

		<label class="field">
			<span class="label">
				Skills (vírgula) — usadas no <strong>% de aderência</strong> das vagas
			</span>
			<Input bind:value={skillsText} placeholder="Go, React, PostgreSQL, Docker…" />
		</label>

		{#if savedSkills.length}
			<div class="skills">
				{#each savedSkills as s (s)}<span class="skill">{s}</span>{/each}
			</div>
		{/if}

		{#if error}<p class="error">{error}</p>{/if}
		<div class="actions">
			{#if savedAt}<span class="saved">Salvo ✓</span>{/if}
			<Button type="submit" disabled={saving}>{saving ? 'Salvando…' : 'Salvar CV'}</Button>
		</div>
	</form>

	<p class="hint">
		As skills viram a base do match com as vagas. Em breve: experiências, adaptação por vaga e
		export em PDF/DOCX.
	</p>
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
	.panel {
		background-color: var(--color-surface);
		border: var(--border-width-1) solid var(--color-border);
		border-radius: var(--radius-lg);
		box-shadow: var(--shadow-sm);
	}
	.form {
		padding: var(--space-6);
		display: flex;
		flex-direction: column;
		gap: var(--space-4);
		max-width: var(--max-prose);
	}
	.grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
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
		min-height: 84px;
	}
	textarea:focus {
		outline: none;
		border-color: var(--color-primary);
		box-shadow: var(--shadow-focus);
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
	}
	.saved {
		font-size: var(--text-sm);
		color: var(--color-success-text);
	}
	.error {
		color: var(--color-danger-text);
		font-size: var(--text-sm);
		margin: 0;
	}
	.hint {
		margin: var(--space-4) 0 0;
		font-size: var(--text-sm);
		color: var(--color-text-muted);
		max-width: var(--max-prose);
	}
</style>
