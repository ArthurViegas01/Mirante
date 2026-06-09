<script>
	import Button from '$lib/components/Button.svelte';
	import Input from '$lib/components/Input.svelte';
	import Select from '$lib/components/Select.svelte';
	import StatusBadge from '$lib/components/StatusBadge.svelte';
	import Sparkline from '$lib/components/Sparkline.svelte';
	import ProgressBar from '$lib/components/ProgressBar.svelte';
	import { api } from '$lib/api.js';
	import { monitor } from '$lib/stores/monitor.svelte.js';
	import {
		svcVariant,
		svcLabel,
		uptimeVariant,
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

	// Inline detail of the expanded service (folded in from the old Monitor page).
	let expandedId = $state('');
	let detail = $state(null);
	let detailLoading = $state(false);

	// Group services by camada in display order; empty camada falls last.
	let groups = $derived.by(() => {
		const map = {};
		for (const s of services) (map[s.camada || ''] ??= []).push(s);
		return [...CAMADA_ORDER, '']
			.filter((k) => map[k]?.length)
			.map((k) => ({ camada: k, items: map[k] }));
	});

	let points = $derived(
		detail ? detail.checks.slice().reverse().map((c) => c.latency_ms || 0) : []
	);

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

	async function loadDetail(id) {
		detail = null;
		detailLoading = true;
		try {
			detail = await api(`/api/services/${id}`);
		} catch (e) {
			error = e.message;
		} finally {
			detailLoading = false;
		}
	}

	function toggle(s) {
		if (expandedId === s.id) {
			expandedId = '';
			detail = null;
		} else {
			expandedId = s.id;
			loadDetail(s.id);
		}
	}

	function pct(u) {
		return u.samples > 0 ? `${(u.up_ratio * 100).toFixed(1)}%` : 'sem dados';
	}

	async function toggleEnabled(s) {
		try {
			await api(`/api/services/${s.id}/enabled`, { method: 'POST', body: { enabled: !s.enabled } });
			await load();
			if (expandedId === s.id) await loadDetail(s.id);
		} catch (e) {
			error = e.message;
		}
	}

	async function removeSvc(s) {
		if (!confirm(`Excluir o serviço "${s.nome}" e seu histórico?`)) return;
		try {
			await api(`/api/services/${s.id}`, { method: 'DELETE' });
			if (expandedId === s.id) {
				expandedId = '';
				detail = null;
			}
			await load();
		} catch (e) {
			error = e.message;
		}
	}

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
		<Button size="sm" variant="secondary" onclick={() => (showForm = !showForm)}>
			{showForm ? 'Cancelar' : 'Adicionar serviço'}
		</Button>
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
							<li class:open={expandedId === s.id}>
								<button class="svc" onclick={() => toggle(s)} aria-expanded={expandedId === s.id}>
									<span class="svc-main">
										<span class="chevron" class:rot={expandedId === s.id} aria-hidden="true">›</span>
										<span class="svc-nome">{s.nome}</span>
										{#if s.provider}<span class="svc-prov">{s.provider}</span>{/if}
									</span>
									<StatusBadge
										status={s.enabled ? svcVariant(s.current_status) : 'info'}
										label={s.enabled ? svcLabel(s.current_status) : 'Pausado'}
									/>
								</button>

								{#if expandedId === s.id}
									<div class="detail">
										{#if detailLoading && !detail}
											<p class="muted small">Carregando…</p>
										{:else if detail}
											<p class="mono target">{detail.service.kind} · {detail.service.target}</p>
											<Sparkline {points} />
											<div class="uptime">
												{#each [['24h', detail.uptime_24h], ['7d', detail.uptime_7d], ['30d', detail.uptime_30d]] as [label, u] (label)}
													<div class="up">
														<div class="up-head"><span>{label}</span><span class="mono">{pct(u)}</span></div>
														<ProgressBar value={u.up_ratio} variant={uptimeVariant(u.up_ratio)} />
													</div>
												{/each}
											</div>
											<div class="detail-actions">
												<Button size="sm" variant="secondary" onclick={() => toggleEnabled(s)}>
													{s.enabled ? 'Pausar' : 'Retomar'}
												</Button>
												<button class="del" onclick={() => removeSvc(s)}>Excluir</button>
											</div>
										{/if}
									</div>
								{/if}
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
	.svcs li {
		border: var(--border-width-1) solid var(--color-border);
		border-radius: var(--radius-md);
		overflow: hidden;
	}
	.svcs li.open {
		border-color: var(--color-border-strong);
	}
	.svc {
		width: 100%;
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: var(--space-3);
		padding: 10px 12px;
		background: none;
		border: none;
		font: inherit;
		text-align: left;
		cursor: pointer;
		color: var(--color-text);
	}
	.svc:hover {
		background-color: var(--color-surface-sunken);
	}
	.svc-main {
		display: flex;
		align-items: baseline;
		gap: var(--space-2);
		min-width: 0;
	}
	.chevron {
		font-family: var(--font-mono);
		color: var(--color-text-muted);
		transition: transform var(--dur-fast) var(--ease-out);
	}
	.chevron.rot {
		transform: rotate(90deg);
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
	.detail {
		padding: var(--space-3) var(--space-4) var(--space-4);
		border-top: var(--border-width-1) solid var(--color-divider);
		background-color: var(--color-surface-sunken);
		display: flex;
		flex-direction: column;
		gap: var(--space-3);
	}
	.target {
		margin: 0;
		font-size: 12px;
		color: var(--color-text-muted);
		word-break: break-all;
	}
	.mono {
		font-family: var(--font-mono);
	}
	.uptime {
		display: flex;
		flex-direction: column;
		gap: var(--space-2);
	}
	.up-head {
		display: flex;
		justify-content: space-between;
		font-size: 12px;
		color: var(--color-text-secondary);
		margin-bottom: 4px;
	}
	.detail-actions {
		display: flex;
		align-items: center;
		gap: var(--space-3);
		margin-top: var(--space-1);
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
	.muted {
		color: var(--color-text-secondary);
		font-size: var(--text-sm);
	}
	.muted.small {
		font-size: 12px;
		margin: 0;
	}
	.error {
		color: var(--color-danger-text);
		font-size: var(--text-sm);
		grid-column: 1 / -1;
		margin: 0;
	}
</style>
