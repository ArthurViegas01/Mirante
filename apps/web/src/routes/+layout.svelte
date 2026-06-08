<script>
	import '../app.css';
	import { onMount } from 'svelte';
	import Sidebar from '$lib/components/Sidebar.svelte';
	import ThemeToggle from '$lib/components/ThemeToggle.svelte';
	import NotificationCenter from '$lib/components/NotificationCenter.svelte';
	import { session } from '$lib/stores/session.svelte.js';
	import { api, setCsrf } from '$lib/api.js';
	import { connectMonitorStream } from '$lib/sse.js';
	import { monitor } from '$lib/stores/monitor.svelte.js';

	let { children } = $props();

	onMount(async () => {
		// Populate the session if a valid cookie already exists (ignore 401).
		try {
			const me = await api('/api/auth/me');
			session.user = me.user;
			if (me.csrf_token) {
				setCsrf(me.csrf_token);
				session.csrf = me.csrf_token;
			}
			connectMonitorStream();
			monitor.loadAlerts().catch(() => {});
		} catch (e) {
			/* not logged in */
		}
	});
</script>

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
	@media (max-width: 720px) {
		.app {
			grid-template-columns: 1fr;
		}
	}
</style>
