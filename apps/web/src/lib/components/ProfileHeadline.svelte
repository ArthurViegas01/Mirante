<script>
	import { onMount } from 'svelte';
	import Button from '$lib/components/Button.svelte';
	import Input from '$lib/components/Input.svelte';
	import { api } from '$lib/api.js';

	let profile = $state(null);
	let editing = $state(false);
	let titulo = $state('');
	let tituloAlvo = $state('');
	let saving = $state(false);
	let error = $state('');

	let hasProfile = $derived(!!(profile && (profile.titulo || profile.titulo_alvo)));

	async function load() {
		try {
			profile = await api('/api/profile');
		} catch {
			/* header is non-critical — ignore */
		}
	}
	onMount(load);

	function openEdit() {
		titulo = profile?.titulo ?? '';
		tituloAlvo = profile?.titulo_alvo ?? '';
		error = '';
		editing = true;
	}

	async function save(e) {
		e.preventDefault();
		saving = true;
		error = '';
		try {
			profile = await api('/api/profile', {
				method: 'PUT',
				body: { titulo, titulo_alvo: tituloAlvo }
			});
			editing = false;
		} catch (e) {
			error = e.message;
		} finally {
			saving = false;
		}
	}
</script>

{#if editing}
	<form class="edit" onsubmit={save}>
		<Input label="Profissão atual" bind:value={titulo} placeholder="Dev Backend Pleno" />
		<Input label="Profissão almejada" bind:value={tituloAlvo} placeholder="Staff Engineer" />
		<div class="edit-actions">
			<Button size="sm" type="submit" disabled={saving}>{saving ? '…' : 'Salvar'}</Button>
			<button type="button" class="link" onclick={() => (editing = false)}>cancelar</button>
		</div>
		{#if error}<p class="error">{error}</p>{/if}
	</form>
{:else if hasProfile}
	<p class="headline">
		{#if profile.titulo_alvo}
			<span class="aim">🎯 Mirando <strong>{profile.titulo_alvo}</strong></span>
		{/if}
		{#if profile.titulo}
			<span class="now">{profile.titulo_alvo ? '· hoje ' : ''}{profile.titulo}</span>
		{/if}
		<button class="link" onclick={openEdit}>editar</button>
	</p>
{:else}
	<button class="link prompt" onclick={openEdit}>+ defina sua profissão (atual e almejada)</button>
{/if}

<style>
	.headline {
		margin: var(--space-2) 0 0;
		font-size: var(--text-sm);
		color: var(--color-text-secondary);
		display: flex;
		flex-wrap: wrap;
		align-items: baseline;
		gap: var(--space-2);
	}
	.aim {
		font-family: var(--font-serif);
		font-style: italic;
		color: var(--color-text);
	}
	.aim strong {
		font-style: normal;
		font-weight: var(--weight-semibold);
		color: var(--color-accent);
	}
	.now {
		color: var(--color-text-muted);
	}
	.link {
		background: none;
		border: none;
		padding: 0;
		font-size: var(--text-sm);
		color: var(--color-text-muted);
		cursor: pointer;
		text-decoration: underline;
		text-underline-offset: 2px;
	}
	.link:hover {
		color: var(--color-text);
	}
	.prompt {
		margin-top: var(--space-2);
		color: var(--color-link);
	}
	.edit {
		margin-top: var(--space-3);
		display: flex;
		flex-wrap: wrap;
		align-items: flex-end;
		gap: var(--space-3);
	}
	.edit-actions {
		display: flex;
		align-items: center;
		gap: var(--space-2);
	}
	.error {
		color: var(--color-danger-text);
		font-size: var(--text-sm);
		margin: 0;
		width: 100%;
	}
</style>
