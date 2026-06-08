<script>
	import '../app.css';
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import Sidebar from '$lib/components/Sidebar.svelte';
	import ThemeToggle from '$lib/components/ThemeToggle.svelte';
	import NotificationCenter from '$lib/components/NotificationCenter.svelte';
	import { session } from '$lib/stores/session.svelte.js';
	import { api, setCsrf } from '$lib/api.js';
	import { connectMonitorStream } from '$lib/sse.js';
	import { monitor } from '$lib/stores/monitor.svelte.js';

	let { children } = $props();

	let authChecked = $state(false);
	let onLogin = $derived($page.url.pathname === '/login');

	onMount(async () => {
		// Populate the session if a valid cookie already exists (ignore 401).
		try {
			const me = await api('/api/auth/me');
			session.user = me.user;
			if (me.csrf_token) {
				setCsrf(me.csrf_token);
				session.csrf = me.csrf_token;
			}
		} catch (e) {
			/* not logged in */
		} finally {
			authChecked = true;
		}
	});

	// Guard: bounce logged-out visitors to /login, and logged-in visitors away from it.
	$effect(() => {
		if (!authChecked) return;
		if (!session.authenticated && !onLogin) {
			goto('/login');
		} else if (session.authenticated && onLogin) {
			goto('/projetos');
		}
	});

	// Start the live monitor stream once authenticated — whether the session came
	// from an existing cookie or a fresh login (so no reload is needed).
	let streamStarted = false;
	$effect(() => {
		if (session.authenticated && !streamStarted) {
			streamStarted = true;
			connectMonitorStream();
			monitor.loadAlerts().catch(() => {});
		}
	});
</script>

{#if !authChecked}
	<div class="boot"><span>Carregando…</span></div>
{:else if onLogin}
	<main class="auth-shell">{@render children()}</main>
{:else if session.authenticated}
	<div class="app">
		<Sidebar />
		<div class="main">
			<header class="topbar">
				<NotificationCenter />
				<ThemeToggle />
			</header>
			<main class="content">
				{@render children()}
			</main>
		</div>
	</div>
{:else}
	<div class="boot"><span>Redirecionando…</span></div>
{/if}

<style>
	.boot {
		display: grid;
		place-items: center;
		min-height: 100dvh;
		background-color: var(--color-bg);
		color: var(--color-text-muted);
		font-family: var(--font-mono);
		font-size: var(--text-sm);
	}
	.auth-shell {
		display: grid;
		place-items: center;
		min-height: 100dvh;
		padding: var(--space-8) var(--space-6);
		background-color: var(--color-bg);
	}
	.app {
		display: grid;
		grid-template-columns: var(--sidebar-width) 1fr;
		min-height: 100dvh;
	}
	.main {
		display: flex;
		flex-direction: column;
		background-color: var(--color-bg);
		min-width: 0;
	}
	.topbar {
		display: flex;
		justify-content: flex-end;
		align-items: center;
		gap: var(--space-3);
		padding: var(--space-4) var(--space-6);
		border-bottom: var(--border-width-1) solid var(--color-divider);
	}
	.content {
		flex: 1;
		padding: var(--space-8) var(--space-6);
		max-width: var(--max-canvas);
		width: 100%;
	}
	@media (max-width: 720px) {
		.app {
			grid-template-columns: 1fr;
		}
	}
</style>
