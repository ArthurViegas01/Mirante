<script>
	import { goto } from '$app/navigation';
	import Button from '$lib/components/Button.svelte';
	import Input from '$lib/components/Input.svelte';
	import BrandMark from '$lib/components/BrandMark.svelte';
	import { api, setCsrf } from '$lib/api.js';
	import { session } from '$lib/stores/session.svelte.js';

	let email = $state('');
	let password = $state('');
	let error = $state('');
	let loading = $state(false);

	async function submit(e) {
		e.preventDefault();
		loading = true;
		error = '';
		try {
			const res = await api('/api/auth/login', { method: 'POST', body: { email, password } });
			setCsrf(res.csrf_token);
			session.csrf = res.csrf_token;
			const me = await api('/api/auth/me');
			session.user = me.user;
			goto('/');
		} catch (err) {
			error = err.message || 'Falha no login';
		} finally {
			loading = false;
		}
	}
</script>

<section class="auth-card">
	<div class="brand"><BrandMark size={26} /></div>
	<div class="intro">
		<h1>Bem-vindo de volta</h1>
		<p class="lead">Entre para abrir sua central de comando.</p>
	</div>
	<form onsubmit={submit}>
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
			autocomplete="current-password"
			bind:value={password}
			required
		/>
		{#if error}<p class="alert" role="alert">{error}</p>{/if}
		<Button type="submit" full disabled={loading}>{loading ? 'Entrando…' : 'Entrar'}</Button>
	</form>
</section>

<style>
	.auth-card {
		width: 100%;
		max-width: 400px;
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
</style>
