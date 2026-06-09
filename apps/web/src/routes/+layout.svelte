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
	import { connectMonitorStream, disconnectMonitorStream } from '$lib/sse.js';
	import { monitor } from '$lib/stores/monitor.svelte.js';

	let { children } = $props();

	// `ready` flips once the initial /me probe settles, so we don't flash the app
	// shell (or redirect) before we know whether a session exists.
	let ready = $state(false);

	let isLogin = $derived($page.url.pathname === '/login');
	let authed = $derived(session.authenticated);

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
			ready = true;
		}
	});

	// Auth guard: logged out → /login; logged in but sitting on /login → app.
	$effect(() => {
		if (!ready) return;
		if (!authed && !isLogin) goto('/login');
		else if (authed && isLogin) goto('/projetos');
	});

	// Monitor stream follows auth: connect on login, drop on logout. Reactive so
	// it also fires after an in-app login (the layout never remounts).
	$effect(() => {
		if (authed) {
			connectMonitorStream();
			monitor.loadAlerts().catch(() => {});
		} else {
			disconnectMonitorStream();
		}
	});
</script>

{#if !ready}
	<div class="boot" aria-hidden="true"></div>
{:else if isLogin}
	<div class="auth-shell">
		{@render children()}
	</div>
{:else if authed}
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
	<!-- Logged out on a protected route: redirecting to /login. -->
	<div class="boot" aria-hidden="true"></div>
{/if}

<style>
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
	.boot {
		min-height: 100dvh;
		background-color: var(--color-bg);
	}
	.auth-shell {
		display: flex;
		flex-direction: column;
		justify-content: center;
		min-height: 100dvh;
		background-color: var(--color-bg);
		padding: var(--space-8) var(--space-6);
	}
	@media (max-width: 720px) {
		.app {
			grid-template-columns: 1fr;
		}
	}
</style>
