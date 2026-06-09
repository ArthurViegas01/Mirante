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
	let experiences = $state([]);
	let education = $state([]);

	let loading = $state(true);
	let saving = $state(false);
	let error = $state('');
	let saved = $state(false);

	let importText = $state('');
	let importing = $state(false);
	let importError = $state('');

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
			experiences = (p.experiences ?? []).map((e) => ({
				empresa: e.empresa,
				cargo: e.cargo,
				inicio: e.inicio,
				fim: e.fim,
				descricao: e.descricao
			}));
			education = (p.education ?? []).map((e) => ({
				instituicao: e.instituicao,
				curso: e.curso,
				inicio: e.inicio,
				fim: e.fim
			}));
		} catch (e) {
			error = e.message;
		} finally {
			loading = false;
		}
	}

	onMount(load);

	async function importCV() {
		if (!importText.trim()) return;
		importing = true;
		importError = '';
		try {
			const d = await api('/api/cv/import', { method: 'POST', body: { text: importText } });
			nome = d.nome || nome;
			titulo = d.titulo || titulo;
			tituloAlvo = d.titulo_alvo || tituloAlvo;
			resumo = d.resumo || resumo;
			if (d.skills?.length) skillsText = d.skills.join(', ');
			if (d.experiences?.length) {
				experiences = d.experiences.map((e) => ({
					empresa: e.empresa,
					cargo: e.cargo,
					inicio: e.inicio,
					fim: e.fim,
					descricao: e.descricao
				}));
			}
			if (d.education?.length) {
				education = d.education.map((e) => ({
					instituicao: e.instituicao,
					curso: e.curso,
					inicio: e.inicio,
					fim: e.fim
				}));
			}
			importText = '';
		} catch (e) {
			importError = e.message;
		} finally {
			importing = false;
		}
	}

	const addExperience = () =>
		experiences.push({ empresa: '', cargo: '', inicio: '', fim: '', descricao: '' });
	const removeExperience = (i) => experiences.splice(i, 1);
	const addEducation = () => education.push({ instituicao: '', curso: '', inicio: '', fim: '' });
	const removeEducation = (i) => education.splice(i, 1);

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
			const p = await api('/api/cv', {
				method: 'PUT',
				body: { nome, titulo, titulo_alvo: tituloAlvo, resumo, skills, experiences, education }
			});
			savedSkills = p.skills ?? [];
			skillsText = savedSkills.join(', ');
			saved = true;
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
	<section class="panel import">
		<h2>Importar de um texto</h2>
		<p class="muted">
			Cole seu CV (ou inventário de skills) — a IA estrutura identidade, skills, experiências e
			educação. Revise e salve.
		</p>
		<textarea bind:value={importText} rows="5" placeholder="Cole aqui o texto do seu CV…"></textarea>
		{#if importError}<p class="error">{importError}</p>{/if}
		<div class="actions">
			<Button variant="secondary" onclick={importCV} disabled={importing || !importText.trim()}>
				{importing ? 'Estruturando…' : '✨ Estruturar com IA'}
			</Button>
		</div>
	</section>

	<form onsubmit={save}>
		<section class="panel">
			<h2>Identidade</h2>
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
				<span class="label">Skills (vírgula) — base do <strong>% de aderência</strong> das vagas</span>
				<Input bind:value={skillsText} placeholder="Go, React, PostgreSQL, Docker…" />
			</label>
			{#if savedSkills.length}
				<div class="skills">
					{#each savedSkills as s (s)}<span class="skill">{s}</span>{/each}
				</div>
			{/if}
		</section>

		<section class="panel">
			<div class="panel-head">
				<h2>Experiências</h2>
				<Button size="sm" variant="secondary" onclick={addExperience}>+ Experiência</Button>
			</div>
			{#if experiences.length === 0}
				<p class="muted">Nenhuma experiência. Adicione seus cargos anteriores.</p>
			{/if}
			{#each experiences as exp, i (i)}
				<div class="entry">
					<div class="grid">
						<Input label="Empresa" bind:value={exp.empresa} />
						<Input label="Cargo" bind:value={exp.cargo} />
						<Input label="Início" bind:value={exp.inicio} placeholder="2022" />
						<Input label="Fim" bind:value={exp.fim} placeholder="atual" />
					</div>
					<label class="field">
						<span class="label">O que você fez</span>
						<textarea bind:value={exp.descricao} rows="3" placeholder="Responsabilidades, resultados, stack."></textarea>
					</label>
					<button type="button" class="del" onclick={() => removeExperience(i)}>Remover experiência</button>
				</div>
			{/each}
		</section>

		<section class="panel">
			<div class="panel-head">
				<h2>Educação</h2>
				<Button size="sm" variant="secondary" onclick={addEducation}>+ Formação</Button>
			</div>
			{#if education.length === 0}
				<p class="muted">Nenhuma formação cadastrada.</p>
			{/if}
			{#each education as ed, i (i)}
				<div class="entry">
					<div class="grid">
						<Input label="Instituição" bind:value={ed.instituicao} />
						<Input label="Curso" bind:value={ed.curso} />
						<Input label="Início" bind:value={ed.inicio} placeholder="2016" />
						<Input label="Fim" bind:value={ed.fim} placeholder="2021" />
					</div>
					<button type="button" class="del" onclick={() => removeEducation(i)}>Remover formação</button>
				</div>
			{/each}
		</section>

		{#if error}<p class="error">{error}</p>{/if}
		<div class="actions">
			{#if saved}<span class="ok">Salvo ✓</span>{/if}
			<Button type="submit" disabled={saving}>{saving ? 'Salvando…' : 'Salvar CV'}</Button>
		</div>
	</form>
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
	h2 {
		font-size: var(--text-lg);
		font-weight: var(--weight-medium);
		color: var(--color-text);
		margin: 0 0 var(--space-4);
	}
	.muted {
		color: var(--color-text-secondary);
		font-size: var(--text-sm);
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
	textarea {
		font-family: var(--font-sans);
		font-size: 13px;
		padding: 10px;
		background-color: var(--color-surface);
		color: var(--color-text);
		border: var(--border-width-1) solid var(--color-border);
		border-radius: var(--radius-md);
		resize: vertical;
		min-height: 72px;
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
	.entry {
		display: flex;
		flex-direction: column;
		gap: var(--space-3);
		border: var(--border-width-1) solid var(--color-divider);
		border-radius: var(--radius-md);
		padding: var(--space-4);
	}
	.del {
		align-self: flex-start;
		background: none;
		border: none;
		color: var(--color-text-muted);
		font-size: var(--text-sm);
		cursor: pointer;
		padding: 0;
	}
	.del:hover {
		color: var(--color-danger-text);
	}
	.actions {
		display: flex;
		align-items: center;
		justify-content: flex-end;
		gap: var(--space-3);
		max-width: var(--max-prose);
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
</style>
