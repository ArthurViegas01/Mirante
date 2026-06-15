<script>
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import Button from '$lib/components/Button.svelte';
	import Input from '$lib/components/Input.svelte';
	import BrandMark from '$lib/components/BrandMark.svelte';
	import { api } from '$lib/api.js';

	// The token rides in the link we e-mailed: /reset-password?token=…
	let token = $derived($page.url.searchParams.get('token') ?? '');

	let password = $state('');
	let confirm = $state('');
	let error = $state('');
	let loading = $state(false);
	let done = $state(false);

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
			await api('/api/auth/reset-password', { method: 'POST', body: { token, password } });
			done = true;
		} catch (err) {
			error = err.message || 'Não foi possível redefinir a senha';
		} finally {
			loading = false;
		}
	}
</script>

<section class="auth-card">
	<div class="brand"><BrandMark size={26} /></div>
	{#if done}
		<div class="intro">
			<h1>Senha redefinida</h1>
			<p class="lead">Sua senha foi atualizada e suas sessões foram encerradas. Entre com a nova senha.</p>
		</div>
		<Button full onclick={() => goto('/login')}>Ir para o login</Button>
	{:else if !token}
		<div class="intro">
			<h1>Link inválido</h1>
			<p class="lead">Este link de redefinição está incompleto ou expirou. Solicite um novo.</p>
		</div>
		<a class="back" href="/forgot-password">Solicitar novo link</a>
	{:else}
		<div class="intro">
			<p class="eyebrow">Recuperação de acesso</p>
			<h1>Crie uma nova senha</h1>
			<p class="lead">Escolha uma nova senha para sua conta.</p>
		</div>
		<form onsubmit={submit}>
			<Input
				label="Nova senha"
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
			<Button type="submit" full disabled={loading}>{loading ? 'Salvando…' : 'Redefinir senha'}</Button>
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
