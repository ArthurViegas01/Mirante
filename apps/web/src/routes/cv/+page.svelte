<script>
	import { onMount } from 'svelte';
	import Button from '$lib/components/Button.svelte';
	import Input from '$lib/components/Input.svelte';
	import Textarea from '$lib/components/Textarea.svelte';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import Skeleton from '$lib/components/Skeleton.svelte';
	import { api } from '$lib/api.js';
	import { toasts } from '$lib/stores/toast.svelte.js';

	let nome = $state('');
	let titulo = $state('');
	let tituloAlvo = $state('');
	let contato = $state('');
	let resumo = $state('');
	let savedSkills = $state([]);
	let experiences = $state([]);
	let education = $state([]);
	// Identity is owned by Meu Perfil, so on save we re-fetch the latest identity
	// instead of echoing this page's snapshot (avoids clobbering identity edited
	// elsewhere). The exception is a CV import, which intentionally rewrites it.
	let importedIdentity = $state(false);

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
			contato = p.contato ?? '';
			resumo = p.resumo ?? '';
			savedSkills = p.skills ?? [];
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
			toasts.error(e.message);
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
			contato = d.contato || contato;
			resumo = d.resumo || resumo;
			if (d.skills?.length) savedSkills = d.skills;
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
			importedIdentity = true;
			toasts.success('CV estruturado pela IA');
		} catch (e) {
			importError = e.message;
			toasts.error(e.message);
		} finally {
			importing = false;
		}
	}

	const addExperience = () =>
		experiences.push({ empresa: '', cargo: '', inicio: '', fim: '', descricao: '' });
	const removeExperience = (i) => experiences.splice(i, 1);
	const addEducation = () => education.push({ instituicao: '', curso: '', inicio: '', fim: '' });
	const removeEducation = (i) => education.splice(i, 1);

	async function persist() {
		// PUT /api/cv is a full replace, but identity lives in Meu Perfil. Unless an
		// import just rewrote it on purpose, re-fetch the freshest identity right
		// before saving so this page never overwrites identity with a stale snapshot
		// (e.g. edited in Meu Perfil in another tab). We only own experiences/education.
		let id = { nome, titulo, titulo_alvo: tituloAlvo, contato, resumo, skills: savedSkills };
		if (!importedIdentity) {
			const fresh = await api('/api/profile');
			id = {
				nome: fresh.nome ?? '',
				titulo: fresh.titulo ?? '',
				titulo_alvo: fresh.titulo_alvo ?? '',
				contato: fresh.contato ?? '',
				resumo: fresh.resumo ?? '',
				skills: fresh.skills ?? []
			};
		}
		const p = await api('/api/cv', {
			method: 'PUT',
			body: { ...id, experiences, education }
		});
		// Reflect the persisted identity in the read-only display and reset the flag.
		nome = p.nome ?? '';
		titulo = p.titulo ?? '';
		tituloAlvo = p.titulo_alvo ?? '';
		contato = p.contato ?? '';
		resumo = p.resumo ?? '';
		savedSkills = p.skills ?? [];
		importedIdentity = false;
		return p;
	}

	async function save(e) {
		e.preventDefault();
		saving = true;
		error = '';
		saved = false;
		try {
			await persist();
			saved = true;
			toasts.success('CV salvo');
		} catch (e) {
			error = e.message;
			toasts.error(e.message);
		} finally {
			saving = false;
		}
	}

	let exporting = $state('');

	async function downloadCV(format) {
		exporting = format;
		error = '';
		try {
			await persist(); // export reflects the current form
			const res = await fetch(`/api/cv/export?format=${format}`, { credentials: 'include' });
			if (!res.ok) throw new Error('falha ao gerar o documento');
			const blob = await res.blob();
			const url = URL.createObjectURL(blob);
			const slug = (nome || 'export').trim().replace(/[^A-Za-z0-9]+/g, '-').replace(/^-|-$/g, '');
			const a = document.createElement('a');
			a.href = url;
			a.download = `CV-${slug || 'export'}.${format}`;
			document.body.appendChild(a);
			a.click();
			a.remove();
			URL.revokeObjectURL(url);
			toasts.success(`${format.toUpperCase()} gerado`);
		} catch (e) {
			error = e.message;
			toasts.error(e.message);
		} finally {
			exporting = '';
		}
	}
</script>

<header class="page-head">
	<div>
		<h1>CV mestre</h1>
	</div>
</header>

{#if loading}
	<div aria-hidden="true">
		{#each Array(3) as _, i (i)}
			<section class="panel sk-panel">
				<Skeleton w="160px" h="18px" />
				<div class="grid">
					{#each Array(3) as __, j (j)}<Skeleton w="100%" h="34px" radius="var(--radius-md)" block />{/each}
				</div>
				<Skeleton w="100%" h="64px" radius="var(--radius-md)" block />
			</section>
		{/each}
	</div>
{:else}
	<section class="panel import">
		<h2>Importar de um texto</h2>
		<p class="muted">
			Cole seu CV (ou inventário de skills): a IA estrutura identidade, skills, experiências e
			educação. Revise e salve.
		</p>
		<Textarea
			bind:value={importText}
			rows={5}
			placeholder="Cole aqui o texto do seu CV…"
			error={importError}
		/>
		<div class="actions">
			<Button variant="secondary" onclick={importCV} disabled={importing || !importText.trim()}>
				{importing ? 'Estruturando…' : '✨ Estruturar com IA'}
			</Button>
		</div>
	</section>

	<form onsubmit={save}>
		<section class="panel">
			<div class="panel-head">
				<h2>Identidade</h2>
				<a class="edit-link" href="/perfil">Editar em Meu Perfil →</a>
			</div>
			{#if nome || titulo || tituloAlvo || contato || resumo || savedSkills.length}
				<dl class="id-list">
					{#if nome}<div><dt>Nome</dt><dd>{nome}</dd></div>{/if}
					{#if titulo}<div><dt>Profissão atual</dt><dd>{titulo}</dd></div>{/if}
					{#if tituloAlvo}<div><dt>Profissão almejada</dt><dd>{tituloAlvo}</dd></div>{/if}
					{#if contato}<div><dt>Contato</dt><dd>{contato}</dd></div>{/if}
					{#if resumo}<div class="wide"><dt>Resumo</dt><dd>{resumo}</dd></div>{/if}
				</dl>
				{#if savedSkills.length}
					<div class="skills">
						{#each savedSkills as s (s)}<span class="skill">{s}</span>{/each}
					</div>
				{/if}
			{:else}
				<p class="muted">Ainda sem identidade. <a href="/perfil">Preencha em Meu Perfil</a>.</p>
			{/if}
		</section>

		<section class="panel">
			<div class="panel-head">
				<h2>Experiências</h2>
				<Button size="sm" variant="secondary" onclick={addExperience}>+ Experiência</Button>
			</div>
			{#if experiences.length === 0}
				<EmptyState
					compact
					title="Nenhuma experiência"
					description="Adicione seus cargos anteriores para compor o CV."
				/>
			{/if}
			{#each experiences as exp, i (i)}
				<div class="entry">
					<div class="grid">
						<Input label="Empresa" bind:value={exp.empresa} />
						<Input label="Cargo" bind:value={exp.cargo} />
						<Input label="Início" bind:value={exp.inicio} placeholder="2022" />
						<Input label="Fim" bind:value={exp.fim} placeholder="atual" />
					</div>
					<Textarea
						label="O que você fez"
						bind:value={exp.descricao}
						rows={3}
						placeholder="Responsabilidades, resultados, stack."
					/>
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
				<EmptyState compact title="Nenhuma formação cadastrada" description="Adicione cursos e graduações." />
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
			<Button variant="secondary" onclick={() => downloadCV('pdf')} disabled={!!exporting || saving}>
				{exporting === 'pdf' ? 'Gerando…' : 'Exportar PDF'}
			</Button>
			<Button variant="secondary" onclick={() => downloadCV('docx')} disabled={!!exporting || saving}>
				{exporting === 'docx' ? 'Gerando…' : 'Exportar DOCX'}
			</Button>
			<Button type="submit" disabled={saving}>{saving ? 'Salvando…' : 'Salvar CV'}</Button>
		</div>
	</form>
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
	.edit-link {
		font-size: var(--text-sm);
		color: var(--color-link);
		text-decoration: none;
		white-space: nowrap;
	}
	.edit-link:hover {
		text-decoration: underline;
	}
	.id-list {
		margin: 0;
		display: flex;
		flex-direction: column;
		gap: var(--space-3);
	}
	.id-list > div {
		display: grid;
		grid-template-columns: 150px 1fr;
		gap: var(--space-3);
	}
	.id-list .wide {
		grid-template-columns: 1fr;
		gap: var(--space-1);
	}
	.id-list dt {
		margin: 0;
		padding-top: 2px;
		font-size: 11px;
		font-weight: var(--weight-medium);
		text-transform: uppercase;
		letter-spacing: 0.04em;
		color: var(--color-text-muted);
	}
	.id-list dd {
		margin: 0;
		font-size: var(--text-sm);
		color: var(--color-text);
	}
	@media (max-width: 520px) {
		.id-list > div {
			grid-template-columns: 1fr;
			gap: var(--space-1);
		}
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
