<script>
	import Button from '$lib/components/Button.svelte';
	import Input from '$lib/components/Input.svelte';
	import Select from '$lib/components/Select.svelte';
	import { api } from '$lib/api.js';
	import {
		formatMoney,
		toCents,
		centsToAmount,
		sumByCurrency,
		formatMonthlyTotals,
		MOEDA_OPTIONS,
		CICLO_OPTIONS
	} from '$lib/money.js';

	let { projectId } = $props();

	let subscriptions = $state([]);
	let services = $state([]);
	let loading = $state(true);
	let error = $state('');

	let showForm = $state(false);
	let editingId = $state(null);
	let nome = $state('');
	let provider = $state('');
	let valor = $state('');
	let moeda = $state('BRL');
	let ciclo = $state('mensal');
	let serviceId = $state('');
	let saving = $state(false);
	let formError = $state('');

	let totalLabel = $derived(formatMonthlyTotals(sumByCurrency(subscriptions)));
	let serviceOptions = $derived([
		{ value: '', label: '— Sem serviço —' },
		...services.map((s) => ({ value: s.id, label: s.nome }))
	]);

	const serviceName = (sid) => services.find((s) => s.id === sid)?.nome ?? '';

	async function load() {
		loading = true;
		error = '';
		try {
			const [subs, svc] = await Promise.all([
				api(`/api/subscriptions?project=${projectId}`),
				api(`/api/services?project_id=${projectId}`)
			]);
			subscriptions = subs.subscriptions;
			services = svc.services;
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

	function resetForm() {
		editingId = null;
		nome = provider = valor = serviceId = '';
		moeda = 'BRL';
		ciclo = 'mensal';
		formError = '';
	}

	function openAdd() {
		resetForm();
		showForm = true;
	}

	function openEdit(s) {
		editingId = s.id;
		nome = s.nome;
		provider = s.provider;
		valor = centsToAmount(s.valor_cents);
		moeda = s.moeda;
		ciclo = s.ciclo;
		serviceId = s.service_id;
		formError = '';
		showForm = true;
	}

	async function submit(e) {
		e.preventDefault();
		saving = true;
		formError = '';
		const body = {
			nome,
			provider,
			valor_cents: toCents(valor),
			moeda,
			ciclo,
			service_id: serviceId
		};
		try {
			if (editingId) {
				await api(`/api/subscriptions/${editingId}`, { method: 'PATCH', body });
			} else {
				await api('/api/subscriptions', { method: 'POST', body: { project_id: projectId, ...body } });
			}
			showForm = false;
			resetForm();
			await load();
		} catch (e) {
			formError = e.message;
		} finally {
			saving = false;
		}
	}

	async function toggleAtivo(s) {
		try {
			await api(`/api/subscriptions/${s.id}`, { method: 'PATCH', body: { ativo: !s.ativo } });
			await load();
		} catch (e) {
			error = e.message;
		}
	}

	async function remove(s) {
		if (!confirm(`Excluir a assinatura "${s.nome}"?`)) return;
		try {
			await api(`/api/subscriptions/${s.id}`, { method: 'DELETE' });
			await load();
		} catch (e) {
			error = e.message;
		}
	}
</script>

<section class="panel">
	<div class="panel-head">
		<div>
			<h2>Custos</h2>
			<p class="total">{totalLabel}{#if totalLabel !== '—'}<span class="total-cap"> · estimativa mensal</span>{/if}</p>
		</div>
		<Button size="sm" variant="secondary" onclick={() => (showForm ? (showForm = false) : openAdd())}>
			{showForm ? 'Cancelar' : 'Nova assinatura'}
		</Button>
	</div>

	{#if showForm}
		<form class="add-form" onsubmit={submit}>
			<Input label="Nome" bind:value={nome} placeholder="Netlify Pro" required />
			<Input label="Provedor" bind:value={provider} placeholder="Netlify" />
			<Input label="Valor" type="number" bind:value={valor} placeholder="0,00" />
			<Select label="Moeda" bind:value={moeda} options={MOEDA_OPTIONS} />
			<Select label="Ciclo" bind:value={ciclo} options={CICLO_OPTIONS} />
			<Select label="Serviço (opcional)" bind:value={serviceId} options={serviceOptions} />
			<div class="add-submit">
				<Button size="sm" type="submit" disabled={saving}>
					{saving ? '…' : editingId ? 'Salvar' : 'Adicionar'}
				</Button>
			</div>
			{#if formError}<p class="error">{formError}</p>{/if}
		</form>
	{/if}

	{#if loading}
		<p class="muted">Carregando…</p>
	{:else if error}
		<p class="error">{error}</p>
	{:else if subscriptions.length === 0}
		<p class="muted">Nenhuma assinatura. Registre os custos recorrentes (hospedagem, banco, domínio…).</p>
	{:else}
		<ul class="subs">
			{#each subscriptions as s (s.id)}
				<li class:inactive={!s.ativo}>
					<div class="sub-main">
						<span class="sub-nome">{s.nome}</span>
						{#if s.provider}<span class="sub-prov">{s.provider}</span>{/if}
						{#if s.service_id && serviceName(s.service_id)}<span class="sub-link">↔ {serviceName(s.service_id)}</span>{/if}
					</div>
					<span class="sub-valor">{formatMoney(s.valor_cents, s.moeda)} / {s.ciclo === 'anual' ? 'ano' : 'mês'}</span>
					<div class="sub-actions">
						<button class="link-btn" onclick={() => toggleAtivo(s)}>{s.ativo ? 'pausar' : 'ativar'}</button>
						<button class="link-btn" onclick={() => openEdit(s)}>editar</button>
						<button class="link-btn danger" onclick={() => remove(s)}>excluir</button>
					</div>
				</li>
			{/each}
		</ul>
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
		align-items: flex-start;
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
	.total {
		margin: 4px 0 0;
		font-size: var(--text-sm);
		color: var(--color-text);
		font-weight: var(--weight-semibold);
	}
	.total-cap {
		color: var(--color-text-muted);
		font-weight: var(--weight-regular);
	}
	.add-form {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
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
	.subs {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: var(--space-2);
	}
	.subs li {
		display: flex;
		align-items: center;
		gap: var(--space-3);
		padding: 10px 12px;
		border: var(--border-width-1) solid var(--color-border);
		border-radius: var(--radius-md);
	}
	.subs li.inactive {
		opacity: 0.55;
	}
	.sub-main {
		display: flex;
		align-items: baseline;
		flex-wrap: wrap;
		gap: var(--space-2);
		min-width: 0;
		flex: 1;
	}
	.sub-nome {
		font-weight: var(--weight-medium);
		color: var(--color-text);
	}
	.sub-prov {
		font-family: var(--font-mono);
		font-size: 11px;
		color: var(--color-text-muted);
	}
	.sub-link {
		font-size: 11px;
		color: var(--color-text-muted);
	}
	.sub-valor {
		font-family: var(--font-mono);
		font-size: 13px;
		color: var(--color-text);
		white-space: nowrap;
	}
	.sub-actions {
		display: flex;
		gap: var(--space-2);
	}
	.link-btn {
		background: none;
		border: none;
		padding: 2px 4px;
		font-size: 12px;
		color: var(--color-text-muted);
		cursor: pointer;
	}
	.link-btn:hover {
		color: var(--color-text);
	}
	.link-btn.danger:hover {
		color: var(--color-danger-text);
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
