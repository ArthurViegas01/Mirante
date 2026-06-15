<script>
	import Button from '$lib/components/Button.svelte';
	import Input from '$lib/components/Input.svelte';
	import BrandMark from '$lib/components/BrandMark.svelte';
	import { api } from '$lib/api.js';

	let email = $state('');
	let error = $state('');
	let loading = $state(false);
	let sent = $state(false);

	async function submit(e) {
		e.preventDefault();
		loading = true;
		error = '';
		try {
			await api('/api/auth/forgot-password', { method: 'POST', body: { email } });
			// The API answers 200 whether or not the address has an account, so the
			// UI confirms without revealing whether the e-mail exists.
			sent = true;
		} catch (err) {
			error = err.message || 'Não foi possível enviar o e-mail';
		} finally {
			loading = false;
		}
	}
</script>

<section class="auth-card">
	<div class="brand"><BrandMark size={26} /></div>
	{#if sent}
		<div class="intro">
			<h1>Verifique seu e-mail</h1>
			<p class="lead">
				Se houver uma conta associada a <strong>{email}</strong>, enviamos um link para redefinir a
				senha. O link expira em 1 hora.
			</p>
		</div>
		<a class="back" href="/login">← Voltar para o login</a>
	{:else}
		<div class="intro">
			<p class="eyebrow">Recuperação de acesso</p>
			<h1>Esqueceu a senha?</h1>
			<p class="lead">Informe seu e-mail e enviaremos um link para você criar uma nova senha.</p>
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
			{#if error}<p class="alert" role="alert">{error}</p>{/if}
			<Button type="submit" full disabled={loading}>{loading ? 'Enviando…' : 'Enviar link'}</Button>
		</form>
		<a class="back" href="/login">← Voltar para o login</a>
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
	.eyebrow {
		margin: 0 0 var(--space-1);
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
