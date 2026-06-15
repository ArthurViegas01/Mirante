<script>
	import { goto } from '$app/navigation';
	import Button from '$lib/components/Button.svelte';
	import Input from '$lib/components/Input.svelte';
	import BrandMark from '$lib/components/BrandMark.svelte';
	import { api, setCsrf } from '$lib/api.js';
	import { session } from '$lib/stores/session.svelte.js';

	let name = $state('');
	let email = $state('');
	let password = $state('');
	let confirm = $state('');
	let error = $state('');
	let loading = $state(false);
	let pending = $state(false);

	let pwError = $derived(password.length > 0 && password.length < 8 ? 'Mínimo de 8 caracteres.' : '');
	let confirmError = $derived(confirm.length > 0 && confirm !== password ? 'As senhas não conferem.' : '');

	async function submit(e) {
		e.preventDefault();
		error = '';
		if (password.length < 8) {
			error = 'A senha precisa ter ao menos 8 caracteres.';
			return;
		}
		if (password !== confirm) {
			error = 'As senhas não conferem.';
			return;
		}
		loading = true;
		try {
			const res = await api('/api/auth/signup', { method: 'POST', body: { email, password, name } });
			if (res.csrf_token) {
				// First account becomes the admin and is logged in immediately.
				setCsrf(res.csrf_token);
				session.csrf = res.csrf_token;
				session.needsSetup = false;
				const me = await api('/api/auth/me');
				session.user = me.user;
				goto('/');
			} else {
				// Created, but awaiting admin activation before it can log in.
				pending = true;
			}
		} catch (err) {
			error = err.message || 'Falha ao criar a conta';
		} finally {
			loading = false;
		}
	}
</script>

<section class="auth-card">
	<div class="brand"><BrandMark size={26} /></div>
	{#if pending}
		<div class="intro">
			<h1>Conta criada</h1>
			<p class="lead">
				Sua conta foi criada e está <strong>aguardando ativação</strong> pelo administrador. Você
				poderá entrar assim que ela for aprovada.
			</p>
		</div>
		<a class="back" href="/login">← Voltar para o login</a>
	{:else}
		<div class="intro">
			<h1>Criar conta</h1>
			<p class="lead">Crie sua conta para acessar o Mirante.</p>
		</div>
		<form onsubmit={submit}>
			<Input label="Nome" name="name" autocomplete="name" bind:value={name} placeholder="Seu nome (opcional)" />
			<Input
				label="E-mail"
				type="email"
				name="email"
				autocomplete="username"
				bind:value={email}
				placeholder="voce@example.com"
				required
			/>
			<Input
				label="Senha"
				type="password"
				name="password"
				autocomplete="new-password"
				bind:value={password}
				hint="Mínimo de 8 caracteres"
				error={pwError}
				required
			/>
			<Input
				label="Confirmar senha"
				type="password"
				name="confirm"
				autocomplete="new-password"
				bind:value={confirm}
				error={confirmError}
				required
			/>
			{#if error}<p class="alert" role="alert">{error}</p>{/if}
			<Button type="submit" full disabled={loading}>{loading ? 'Criando…' : 'Criar conta'}</Button>
		</form>
		<p class="aux">Já tem uma conta? <a href="/login">Entrar</a></p>
	{/if}
</section>

<style>
	.auth-card {
		width: 100%;
		max-width: 420px;
		margin-inline: auto;
		display: flex;
		flex-direction: column;
		gap: var(--space-6);
		padding: var(--space-8);
		background-color: var(--color-surface);
		border: var(--border-width-1) solid var(--color-border);
		border-radius: var(--radius-2xl);
		box-shadow: var(--shadow-lg);
		color: var(--color-text);
		--mark-fill: var(--color-primary);
		--word-size: 22px;
	}
	.intro {
		display: flex;
		flex-direction: column;
		gap: var(--space-1);
	}
	h1 {
		font-size: var(--text-xl);
		font-weight: var(--weight-medium);
		letter-spacing: var(--tracking-snug);
		color: var(--color-text);
		margin: 0;
	}
	.lead {
		font-size: var(--text-sm);
		color: var(--color-text-muted);
		line-height: var(--leading-normal);
		margin: 0;
	}
	form {
		display: flex;
		flex-direction: column;
		gap: var(--space-4);
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
	.aux {
		margin: 0;
		text-align: center;
		font-size: var(--text-sm);
		color: var(--color-text-muted);
	}
	.aux a {
		color: var(--color-text-muted);
		text-decoration: none;
		transition: color var(--dur-fast) var(--ease-out);
	}
	.aux a:hover {
		color: var(--color-text);
		text-decoration: underline;
	}
	.back {
		align-self: center;
		font-size: var(--text-sm);
		color: var(--color-text-muted);
		text-decoration: none;
		transition: color var(--dur-fast) var(--ease-out);
	}
	.back:hover {
		color: var(--color-text);
	}
</style>
