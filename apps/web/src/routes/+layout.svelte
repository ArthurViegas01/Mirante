<script>
	import '../app.css';
	import { onMount } from 'svelte';
	import Sidebar from '$lib/components/Sidebar.svelte';
	import ThemeToggle from '$lib/components/ThemeToggle.svelte';
	import { session } from '$lib/stores/session.svelte.js';
	import { api } from '$lib/api.js';

	let { children } = $props();

	onMount(async () => {
		// Populate the session if a valid cookie already exists (ignore 401).
		try {
			const me = await api('/api/auth/me');
			session.user = me.user;
		} catch (e) {
			/* not logged in */
		}
	});
</script>

<div class="app">
	<Sidebar />
	<div class="main">
		<header class="topbar">
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
