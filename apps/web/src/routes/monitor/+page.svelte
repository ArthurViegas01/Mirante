<script>
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import Button from '$lib/components/Button.svelte';
	import Input from '$lib/components/Input.svelte';
	import Select from '$lib/components/Select.svelte';
	import StatusBadge from '$lib/components/StatusBadge.svelte';
	import { api } from '$lib/api.js';
	import { monitor } from '$lib/stores/monitor.svelte.js';
	import { svcVariant, svcLabel, KIND_OPTIONS } from '$lib/serviceStatus.js';

	let services = $state([]);
	let projects = $state([]);
	let loading = $state(true);
	let error = $state('');
	let showForm = $state(false);

	let projectId = $state('');
	let nome = $state('');
	let kind = $state('http');
	let target = $state('');
	let expectedStatus = $state('2xx');
	let intervalSeconds = $state('60');
	let degradedMs = $state('500');
	let timeoutMs = $state('5000');
	let saving = $state(false);
	let formError = $state('');

	async function load() {
		loading = true;
		error = '';
		try {
			const [svc, proj] = await Promise.all([api('/api/services'), api('/api/projects')]);
			services = svc.services;
			projects = proj.projects;
			if (!projectId && projects.length) projectId = projects[0].id;
		} catch (e) {
			error = e.message;
		} finally {
			loading = false;
		}
	}

	onMount(load);

	// Live: when a transition arrives, patch the matching service's status.
	$effect(() => {
		const ev = monitor.lastEvent;
		if (!ev) return;
		services = services.map((s) => (s.id === ev.service_id ? { ...s, current_status: ev.to } : s));
	});

	function projName(id) {
		return projects.find((p) => p.id === id)?.nome ?? '—';
	}

	const projectOptions = $derived(projects.map((p) => ({ value: p.id, label: p.nome })));

	async function create(e) {
		e.preventDefault();
		saving = true;
		formError = '';
		try {
			await api('/api/services', {
				method: 'POST',
				body: {
					project_id: projectId,
					nome,
					kind,
					target,
					expected_status: expectedStatus,
					interval_seconds: Number(intervalSeconds),
					degraded_threshold_ms: Number(degradedMs),
					timeout_ms: Number(timeoutMs)
				}
			});
			showForm = false;
			nome = target = '';
			await load();
		} catch (e) {
			formError = e.message;
		} finally {
			saving = false;
		}
	}
</script>

<header class="page-head">
	<div>
		<p class="eyebrow">
			Monitor
			<span class="live" class:on={monitor.connected}>{monitor.connected ? 'ao vivo' : 'offline'}</span>
		</p>
		<h1>Monitor</h1>
	</div>
	<Button onclick={() => (showForm = !showForm)} disabled={projects.length === 0}>
		{showForm ? 'Cancelar' : 'Novo serviço'}
	</Button>
</header>

{#if projects.length === 0 && !loading}
	<div class="panel empty">Crie um projeto antes de adicionar serviços ao monitor.</div>
{/if}

{#if showForm}
	<form class="panel form" onsubmit={create}>
		<div class="grid">
			<Select label="Projeto" bind:value={projectId} options={projectOptions} />
			<Input label="Nome" bind:value={nome} required />
			<Select label="Tipo" bind:value={kind} options={KIND_OPTIONS} />
			<Input label={kind === 'http' ? 'Alvo (URL)' : 'Alvo (host:porta)'} bind:value={target} required />
			{#if kind === 'http'}
				<Input label="Status esperado" bind:value={expectedStatus} placeholder="2xx | 200,204" />
			{/if}
			<Input label="Intervalo (s)" type="number" bind:value={intervalSeconds} />
			<Input label="Degradado (ms)" type="number" bind:value={degradedMs} />
			<Input label="Timeout (ms)" type="number" bind:value={timeoutMs} />
		</div>
		{#if formError}<p class="error">{formError}</p>{/if}
		<div class="actions">
			<Button type="submit" disabled={saving}>{saving ? 'Salvando…' : 'Adicionar serviço'}</Button>
		</div>
	</form>
{/if}

{#if loading}
	<p class="muted">Carregando…</p>
{:else if error}
	<p class="error">{error}</p>
{:else if services.length === 0}
	<div class="panel empty">Nenhum serviço monitorado ainda.</div>
{:else}
	<div class="board">
		{#each services as s (s.id)}
			<button class="card" onclick={() => goto(`/monitor/${s.id}`)}>
				<div class="card-head">
					<span class="nome">{s.nome}</span>
					<StatusBadge status={svcVariant(s.current_status)} label={svcLabel(s.current_status)} />
				</div>
				<span class="proj">{projName(s.project_id)}</span>
				<span class="target">{s.target}</span>
				<span class="meta">{s.kind} · {s.interval_seconds}s{s.enabled ? '' : ' · pausado'}</span>
			</button>
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
		display: flex;
		align-items: center;
		gap: var(--space-2);
	}
	.live {
		display: inline-flex;
		align-items: center;
		gap: 5px;
		color: var(--color-text-muted);
	}
	.live::before {
		content: '';
		width: 7px;
		height: 7px;
		border-radius: var(--radius-full);
		background-color: var(--color-text-disabled);
	}
	.live.on {
		color: var(--color-accent);
	}
	.live.on::before {
		background-color: var(--color-live);
		box-shadow: 0 0 0 3px var(--color-live-halo);
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
	}
	.panel {
		background-color: var(--color-surface);
		border: var(--border-width-1) solid var(--color-border);
		border-radius: var(--radius-lg);
		box-shadow: var(--shadow-sm);
	}
	.empty {
		padding: var(--space-8);
		text-align: center;
		color: var(--color-text-muted);
		margin-bottom: var(--space-6);
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
		grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
		gap: var(--space-4);
	}
	.card {
		display: flex;
		flex-direction: column;
		gap: 6px;
		text-align: left;
		padding: 18px 20px;
		background-color: var(--color-surface);
		border: var(--border-width-1) solid var(--color-border);
		border-radius: var(--radius-lg);
		box-shadow: var(--shadow-sm);
		cursor: pointer;
		transition:
			border-color var(--dur-fast) var(--ease-out),
			box-shadow var(--dur-fast) var(--ease-out),
			transform var(--dur-fast) var(--ease-out);
	}
	.card:hover {
		border-color: var(--color-border-strong);
		box-shadow: var(--shadow-md);
		transform: translateY(-2px);
	}
	.card:focus-visible {
		outline: none;
		box-shadow: var(--shadow-focus);
	}
	.card-head {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: var(--space-2);
	}
	.nome {
		font-weight: var(--weight-semibold);
		color: var(--color-text);
	}
	.proj {
		font-size: var(--text-sm);
		color: var(--color-text-secondary);
	}
	.target {
		font-family: var(--font-mono);
		font-size: 12px;
		color: var(--color-text-muted);
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.meta {
		font-family: var(--font-mono);
		font-size: 11px;
		color: var(--color-text-disabled);
	}
</style>
