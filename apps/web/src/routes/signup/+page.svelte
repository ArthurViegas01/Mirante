<script>
	import { goto } from '$app/navigation';
	import Button from '$lib/components/Button.svelte';
	import Input from '$lib/components/Input.svelte';
	import { api, setCsrf } from '$lib/api.js';
	import { session } from '$lib/stores/session.svelte.js';

	let name = $state('');
	let email = $state('');
	let password = $state('');
	let confirm = $state('');
	let error = $state('');
	let loading = $state(false);

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
			const res = await api('/api/auth/signup', {
				method: 'POST',
				body: { email, password, name }
			});
			setCsrf(res.csrf_token);
			session.csrf = res.csrf_token;
			session.needsSetup = false;
			const me = await api('/api/auth/me');
			session.user = me.user;
			goto('/projetos');
		} catch (err) {
			error = err.message || 'Falha ao criar o acesso';
		} finally {
			loading = false;
		}
	}
</script>

<section class="signup">
	<p class="eyebrow">Mirante</p>
	<h1>Configurar acesso</h1>
	<p class="lead">Esta é a primeira vez por aqui. Crie a conta do dono para começar.</p>
	<form onsubmit={submit}>
		<Input label="Nome" bind:value={name} placeholder="Seu nome (opcional)" />
		<Input label="E-mail" type="email" bind:value={email} placeholder="voce@example.com" required />
		<Input label="Senha" type="password" bind:value={password} placeholder="Mín. 8 caracteres" required />
		<Input label="Confirmar senha" type="password" bind:value={confirm} required />
		{#if error}<p class="error" role="alert">{error}</p>{/if}
		<Button type="submit" disabled={loading}>{loading ? 'Criando…' : 'Criar acesso'}</Button>
	</form>
	<p class="footer">by Lumni</p>
</section>

<style>
	.signup {
		max-width: 360px;
		margin-inline: auto;
	}
	.eyebrow {
		font-family: var(--font-mono);
		font-size: var(--text-xs);
		letter-spacing: var(--tracking-eyebrow);
		text-transform: uppercase;
		color: var(--color-text-muted);
		margin: 0 0 var(--space-2);
	}
	h1 {
		font-size: var(--text-2xl);
		font-weight: var(--weight-medium);
		letter-spacing: var(--tracking-snug);
		color: var(--color-text);
		margin: 0 0 var(--space-2);
	}
	.lead {
		font-size: var(--text-sm);
		color: var(--color-text-muted);
		margin: 0 0 var(--space-6);
	}
	form {
		display: flex;
		flex-direction: column;
		gap: var(--space-4);
	}
	.error {
		color: var(--color-danger-text);
		font-size: var(--text-sm);
		margin: 0;
	}
	.footer {
		margin-top: var(--space-8);
		font-family: var(--font-mono);
		font-size: var(--text-xs);
		color: var(--color-text-muted);
		text-align: center;
	}
</style>
