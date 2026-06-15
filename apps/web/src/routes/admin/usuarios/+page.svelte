<script>
	import { onMount } from 'svelte';
	import Button from '$lib/components/Button.svelte';
	import Input from '$lib/components/Input.svelte';
	import Skeleton from '$lib/components/Skeleton.svelte';
	import { api } from '$lib/api.js';
	import { toasts } from '$lib/stores/toast.svelte.js';
	import { confirm } from '$lib/stores/confirm.svelte.js';
	import { session } from '$lib/stores/session.svelte.js';

	let users = $state([]);
	let loading = $state(true);
	let loadError = $state('');

	// New-user form (admin creates an already-active account).
	let email = $state('');
	let nome = $state('');
	let senha = $state('');
	let adminRole = $state(false);
	let creating = $state(false);
	let formError = $state('');

	const statusLabel = { active: 'Ativo', pending: 'Pendente', disabled: 'Desativado' };
	const roleLabel = { admin: 'Admin', user: 'Usuário' };

	async function load() {
		loading = true;
		loadError = '';
		try {
			const res = await api('/api/admin/users');
			users = res.users ?? [];
		} catch (e) {
			loadError = e.message || 'Não foi possível carregar os usuários';
		} finally {
			loading = false;
		}
	}
	onMount(load);

	function fmtDate(iso) {
		if (!iso) return '—';
		const d = new Date(iso);
		return Number.isNaN(d.getTime()) ? '—' : d.toLocaleDateString('pt-BR');
	}

	async function activate(u) {
		try {
			await api(`/api/admin/users/${u.id}/activate`, { method: 'POST' });
			toasts.success(`${u.email} ativado.`);
			await load();
		} catch (e) {
			toasts.error(e.message || 'Falha ao ativar.');
		}
	}

	async function deactivate(u) {
		try {
			await api(`/api/admin/users/${u.id}/deactivate`, { method: 'POST' });
			toasts.success(`${u.email} desativado.`);
			await load();
		} catch (e) {
			toasts.error(e.message || 'Falha ao desativar.');
		}
	}

	async function remove(u) {
		const ok = await confirm.ask({
			title: 'Excluir usuário',
			message: `Excluir ${u.email}? Todos os dados dessa conta serão apagados. Esta ação é irreversível.`,
			confirmLabel: 'Excluir',
			danger: true
		});
		if (!ok) return;
		try {
			await api(`/api/admin/users/${u.id}`, { method: 'DELETE' });
			toasts.success(`${u.email} excluído.`);
			await load();
		} catch (e) {
			toasts.error(e.message || 'Falha ao excluir.');
		}
	}

	async function addUser(e) {
		e.preventDefault();
		formError = '';
		if (senha.length < 8) {
			formError = 'A senha precisa ter ao menos 8 caracteres.';
			return;
		}
		creating = true;
		try {
			await api('/api/admin/users', {
				method: 'POST',
				body: { email, password: senha, name: nome, role: adminRole ? 'admin' : 'user' }
			});
			toasts.success(`${email} criado e ativado.`);
			email = '';
			nome = '';
			senha = '';
			adminRole = false;
			await load();
		} catch (e2) {
			formError = e2.message || 'Falha ao criar usuário.';
		} finally {
			creating = false;
		}
	}

	const isSelf = (u) => u.id === session.user?.id;
</script>

<div class="head">
	<h1>Usuários</h1>
	<p class="sub">Aprove novos cadastros e gerencie as contas do Mirante.</p>
</div>

<section class="card">
	<h2>Adicionar usuário</h2>
	<form class="add" onsubmit={addUser}>
		<div class="grid">
			<Input label="Nome" name="nome" bind:value={nome} placeholder="Opcional" />
			<Input label="E-mail" type="email" name="email" bind:value={email} placeholder="pessoa@example.com" required />
			<Input label="Senha" type="password" name="senha" bind:value={senha} hint="Mínimo de 8 caracteres" required />
		</div>
		<label class="check">
			<input type="checkbox" bind:checked={adminRole} />
			<span>Tornar administrador</span>
		</label>
		{#if formError}<p class="alert" role="alert">{formError}</p>{/if}
		<div class="actions-row">
			<Button type="submit" disabled={creating}>{creating ? 'Criando…' : 'Criar e ativar'}</Button>
		</div>
	</form>
</section>

<section class="card">
	<h2>Contas</h2>
	{#if loading}
		<Skeleton lines={4} />
	{:else if loadError}
		<p class="alert" role="alert">{loadError}</p>
	{:else}
		<div class="table-wrap">
			<table>
				<thead>
					<tr>
						<th>Usuário</th>
						<th>Papel</th>
						<th>Status</th>
						<th>Criado</th>
						<th class="right">Ações</th>
					</tr>
				</thead>
				<tbody>
					{#each users as u (u.id)}
						<tr>
							<td>
								<div class="who">
									<span class="nome">{u.name || u.email.split('@')[0]}</span>
									<span class="email">{u.email}</span>
								</div>
							</td>
							<td><span class="role" class:admin={u.role === 'admin'}>{roleLabel[u.role] ?? u.role}</span></td>
							<td><span class="badge {u.status}">{statusLabel[u.status] ?? u.status}</span></td>
							<td class="muted">{fmtDate(u.created_at)}</td>
							<td class="right">
								{#if isSelf(u)}
									<span class="you">você</span>
								{:else}
									<div class="row-actions">
										{#if u.status !== 'active'}
											<Button size="sm" variant="secondary" onclick={() => activate(u)}>Ativar</Button>
										{:else}
											<Button size="sm" variant="ghost" onclick={() => deactivate(u)}>Desativar</Button>
										{/if}
										<Button size="sm" variant="danger" onclick={() => remove(u)}>Excluir</Button>
									</div>
								{/if}
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</section>

<style>
	.head {
		margin-bottom: var(--space-6);
	}
	h1 {
		font-size: var(--text-2xl);
		font-weight: var(--weight-semibold);
		letter-spacing: var(--tracking-snug);
		margin: 0;
		color: var(--color-text);
	}
	.sub {
		margin: var(--space-1) 0 0;
		color: var(--color-text-muted);
		font-size: var(--text-sm);
	}
	.card {
		background-color: var(--color-surface);
		border: var(--border-width-1) solid var(--color-border);
		border-radius: var(--radius-xl);
		padding: var(--space-6);
		margin-bottom: var(--space-6);
	}
	h2 {
		font-size: var(--text-base);
		font-weight: var(--weight-semibold);
		margin: 0 0 var(--space-4);
		color: var(--color-text);
	}
	.add {
		display: flex;
		flex-direction: column;
		gap: var(--space-4);
	}
	.grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
		gap: var(--space-4);
	}
	.check {
		display: inline-flex;
		align-items: center;
		gap: var(--space-2);
		font-size: var(--text-sm);
		color: var(--color-text-secondary);
		cursor: pointer;
	}
	.actions-row {
		display: flex;
		justify-content: flex-end;
	}
	.table-wrap {
		overflow-x: auto;
	}
	table {
		width: 100%;
		border-collapse: collapse;
		font-size: var(--text-sm);
	}
	th {
		text-align: left;
		font-weight: var(--weight-medium);
		color: var(--color-text-muted);
		padding: 0 var(--space-3) var(--space-3);
		border-bottom: var(--border-width-1) solid var(--color-divider);
		white-space: nowrap;
	}
	td {
		padding: var(--space-3);
		border-bottom: var(--border-width-1) solid var(--color-divider);
		vertical-align: middle;
	}
	.right {
		text-align: right;
	}
	.muted {
		color: var(--color-text-muted);
	}
	.who {
		display: flex;
		flex-direction: column;
	}
	.nome {
		color: var(--color-text);
		font-weight: var(--weight-medium);
	}
	.email {
		color: var(--color-text-muted);
		font-size: var(--text-xs);
	}
	.role {
		font-size: var(--text-xs);
		color: var(--color-text-secondary);
	}
	.role.admin {
		color: var(--color-primary);
		font-weight: var(--weight-medium);
	}
	.badge {
		display: inline-block;
		padding: 2px 9px;
		border-radius: var(--radius-full);
		font-size: var(--text-xs);
		font-weight: var(--weight-medium);
	}
	.badge.active {
		background-color: color-mix(in srgb, var(--color-success) 16%, transparent);
		color: var(--color-success);
	}
	.badge.pending {
		background-color: color-mix(in srgb, var(--color-warning) 18%, transparent);
		color: var(--color-warning);
	}
	.badge.disabled {
		background-color: var(--color-surface-sunken);
		color: var(--color-text-muted);
	}
	.row-actions {
		display: inline-flex;
		gap: var(--space-2);
		justify-content: flex-end;
	}
	.you {
		font-size: var(--text-xs);
		color: var(--color-text-muted);
		font-style: italic;
	}
	.alert {
		margin: 0;
		padding: 10px 12px;
		border-radius: var(--radius-md);
		background-color: var(--color-danger-bg);
		color: var(--color-danger-text);
		border: var(--border-width-1) solid color-mix(in srgb, var(--color-danger) 30%, transparent);
		font-size: var(--text-sm);
	}
</style>
