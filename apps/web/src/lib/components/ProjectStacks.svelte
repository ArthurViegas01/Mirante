<script>
	import Button from '$lib/components/Button.svelte';
	import Input from '$lib/components/Input.svelte';
	import Select from '$lib/components/Select.svelte';
	import StatusBadge from '$lib/components/StatusBadge.svelte';
	import { api } from '$lib/api.js';
	import { monitor } from '$lib/stores/monitor.svelte.js';
	import {
		svcVariant,
		svcLabel,
		KIND_OPTIONS,
		CAMADA_OPTIONS,
		CAMADA_ORDER,
		camadaLabel
	} from '$lib/serviceStatus.js';

	let { projectId } = $props();

	let services = $state([]);
	let loading = $state(true);
	let error = $state('');
	let showForm = $state(false);

	let nome = $state('');
	let provider = $state('');
	let camada = $state('');
	let kind = $state('http');
	let target = $state('');
	let saving = $state(false);
	let formError = $state('');

	// Group services by camada in display order; empty camada falls last.
	let groups = $derived.by(() => {
		const map = {};
		for (const s of services) (map[s.camada || ''] ??= []).push(s);
		return [...CAMADA_ORDER, '']
			.filter((k) => map[k]?.length)
			.map((k) => ({ camada: k, items: map[k] }));
	});

	async function load() {
		loading = true;
		error = '';
		try {
			const res = await api(`/api/services?project_id=${projectId}`);
			services = res.services;
		} catch (e) {
			error = e.message;
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		projectId;
		load();
	});

	// Live: patch the matching service's status when a transition arrives.
	$effect(() => {
		const ev = monitor.lastEvent;
		if (!ev) return;
		services = services.map((s) => (s.id === ev.service_id ? { ...s, current_status: ev.to } : s));
	});

	async function create(e) {
		e.preventDefault();
		saving = true;
		formError = '';
		try {
			await api('/api/services', {
				method: 'POST',
				body: { project_id: projectId, nome, provider, camada, kind, target }
			});
			nome = provider = target = '';
			camada = '';
			kind = 'http';
			showForm = false;
			await load();
		} catch (e) {
			formError = e.message;
		} finally {
			saving = false;
		}
	}
</script>

<section class="panel">
	<div class="panel-head">
		<h2>Stacks</h2>
		<div class="head-actions">
			<a class="muted-link" href="/monitor">Monitor →</a>
			<Button size="sm" variant="secondary" onclick={() => (showForm = !showForm)}>
				{showForm ? 'Cancelar' : 'Adicionar serviço'}
			</Button>
		</div>
	</div>

	{#if showForm}
		<form class="add-form" onsubmit={create}>
			<Input label="Nome" bind:value={nome} required />
			<Input label="Provedor" bind:value={provider} placeholder="Netlify, Railway…" />
			<Select label="Camada" bind:value={camada} options={CAMADA_OPTIONS} />
			<Select label="Tipo" bind:value={kind} options={KIND_OPTIONS} />
			<Input
				label={kind === 'http' ? 'Alvo (URL)' : 'Alvo (host:porta)'}
				bind:value={target}
				required
			/>
			<div class="add-submit"><Button size="sm" type="submit" disabled={saving}>{saving ? '…' : 'Adicionar'}</Button></div>
			{#if formError}<p class="error">{formError}</p>{/if}
		</form>
	{/if}

	{#if loading}
		<p class="muted">Carregando…</p>
	{:else if error}
		<p class="error">{error}</p>
	{:else if services.length === 0}
		<p class="muted">Nenhum serviço. Adicione as peças do stack (front, back, banco…) para ver o status ao vivo.</p>
	{:else}
		<div class="groups">
			{#each groups as g (g.camada)}
				<div class="group">
					<p class="camada">{camadaLabel(g.camada)}</p>
					<ul class="svcs">
						{#each g.items as s (s.id)}
							<li>
								<a class="svc" href={`/monitor/${s.id}`}>
									<span class="svc-main">
										<span class="svc-nome">{s.nome}</span>
										{#if s.provider}<span class="svc-prov">{s.provider}</span>{/if}
									</span>
									<StatusBadge
										status={s.enabled ? svcVariant(s.current_status) : 'info'}
										label={s.enabled ? svcLabel(s.current_status) : 'Pausado'}
									/>
								</a>
							</li>
						{/each}
					</ul>
				</div>
			{/each}
		</div>
	{/if}
</section>

<style>
	.panel {
		background-color: var(--color-surface);
		border: var(--border-width-1) solid var(--color-border);
		border-radius: var(--radius-lg);
		box-shadow: var(--shadow-sm);
		padding: var(--space-6);
		margin-bottom: var(--space-5);
	}
	.panel-head {
		display: flex;
		align-items: baseline;
		justify-content: space-between;
		gap: var(--space-4);
		margin-bottom: var(--space-4);
	}
	.head-actions {
		display: flex;
		align-items: center;
		gap: var(--space-3);
	}
	.muted-link {
		font-size: var(--text-sm);
		color: var(--color-link);
		text-decoration: none;
		white-space: nowrap;
	}
	h2 {
		font-size: var(--text-lg);
		font-weight: var(--weight-medium);
		color: var(--color-text);
		margin: 0;
	}
	.add-form {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
		gap: var(--space-3);
		align-items: end;
		border: var(--border-width-1) solid var(--color-divider);
		border-radius: var(--radius-md);
		padding: var(--space-4);
		margin-bottom: var(--space-4);
	}
	.add-submit {
		display: flex;
	}
	.groups {
		display: flex;
		flex-direction: column;
		gap: var(--space-4);
	}
	.camada {
		font-family: var(--font-mono);
		font-size: 11px;
		letter-spacing: 0.06em;
		text-transform: uppercase;
		color: var(--color-text-muted);
		margin: 0 0 var(--space-2);
	}
	.svcs {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: var(--space-2);
	}
	.svc {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: var(--space-3);
		padding: 10px 12px;
		border: var(--border-width-1) solid var(--color-border);
		border-radius: var(--radius-md);
		text-decoration: none;
		transition: border-color var(--dur-fast) var(--ease-out);
	}
	.svc:hover {
		border-color: var(--color-border-strong);
	}
	.svc-main {
		display: flex;
		align-items: baseline;
		gap: var(--space-2);
		min-width: 0;
	}
	.svc-nome {
		font-weight: var(--weight-medium);
		color: var(--color-text);
	}
	.svc-prov {
		font-family: var(--font-mono);
		font-size: 11px;
		color: var(--color-text-muted);
	}
	.muted {
		color: var(--color-text-secondary);
		font-size: var(--text-sm);
	}
	.error {
		color: var(--color-danger-text);
		font-size: var(--text-sm);
		grid-column: 1 / -1;
		margin: 0;
	}
</style>
