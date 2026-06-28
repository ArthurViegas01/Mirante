<script>
	import '../app.css';
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto, onNavigate } from '$app/navigation';
	import Sidebar from '$lib/components/Sidebar.svelte';
	import ThemeToggle from '$lib/components/ThemeToggle.svelte';
	import NotificationCenter from '$lib/components/NotificationCenter.svelte';
	import Toaster from '$lib/components/Toaster.svelte';
	import ConfirmHost from '$lib/components/ConfirmHost.svelte';
	import { session } from '$lib/stores/session.svelte.js';
	import { api, setCsrf, setUnauthorizedHandler } from '$lib/api.js';
	import { connectMonitorStream, disconnectMonitorStream } from '$lib/sse.js';
	import { monitor } from '$lib/stores/monitor.svelte.js';

	let { children } = $props();

	// Cross-page fade via the View Transitions API. Browsers without support fall
	// back to an instant swap; motion itself is clamped by the reduced-motion
	// rules in app.css.
	onNavigate((navigation) => {
		if (!document.startViewTransition) return;
		return new Promise((resolve) => {
			document.startViewTransition(async () => {
				resolve();
				await navigation.complete;
			});
		});
	});

	// `ready` flips once the initial /me probe settles, so we don't flash the app
	// shell (or redirect) before we know whether a session exists.
	let ready = $state(false);

	let path = $derived($page.url.pathname);
	// Public screens, reachable while logged out. Login is the single entry point;
	// the others are linked from it (signup, password recovery).
	const AUTH_ROUTES = ['/login', '/signup', '/forgot-password', '/reset-password'];
	let isAuthRoute = $derived(AUTH_ROUTES.includes(path));
	let isAdminRoute = $derived(path.startsWith('/admin'));
	let authed = $derived(session.authenticated);

	// Mobile drawer state + the current section label shown in the topbar.
	let navOpen = $state(false);
	const SECTION_TITLES = {
		'/': 'Início',
		'/projetos': 'Projetos',
		'/tarefas': 'Tarefas',
		'/custos': 'Custos',
		'/perfil': 'Meu Perfil',
		'/vagas': 'Vagas',
		'/cv': 'Currículo',
		'/candidaturas': 'Candidaturas'
	};
	let pageTitle = $derived(
		SECTION_TITLES[path] ?? (path.startsWith('/projetos/') ? 'Projeto' : 'Mirante')
	);

	onMount(async () => {
		// An expired/revoked session (a 401 on a protected route) drops us back to
		// login. No-ops on the expected 401 from the probe below, since no session
		// is established yet.
		setUnauthorizedHandler(() => {
			if (session.authenticated) {
				session.clear();
				goto('/login');
			}
		});

		// Populate the session if a valid cookie already exists (ignore 401).
		try {
			const me = await api('/api/auth/me');
			session.user = me.user;
			if (me.csrf_token) {
				setCsrf(me.csrf_token);
				session.csrf = me.csrf_token;
			}
		} catch (e) {
			// Not logged in — find out whether the instance still needs its owner.
			try {
				const st = await api('/api/auth/status');
				session.needsSetup = !!st.needs_setup;
			} catch (e2) {
				/* default to login */
			}
		} finally {
			ready = true;
		}
	});

	// Auth guard: anonymous visitors always land on login (the single entry point),
	// from which signup and recovery are linked; logged-in users are kept off the
	// public auth screens.
	$effect(() => {
		if (!ready) return;
		if (authed) {
			if (isAuthRoute) goto('/');
			else if (isAdminRoute && !session.isAdmin) goto('/'); // admin-only area
		} else if (!isAuthRoute) {
			goto('/login');
		}
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
{:else if isAuthRoute}
	<div class="auth-shell">
		{@render children()}
	</div>
{:else if authed}
	<div class="app">
		<Sidebar bind:open={navOpen} />
		<div class="main">
			<header class="topbar">
				<div class="bar-left">
					<button
						class="hamburger"
						onclick={() => (navOpen = true)}
						aria-label="Abrir menu"
						aria-expanded={navOpen}
					>
						<svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" aria-hidden="true">
							<path d="M3 6h18M3 12h18M3 18h18" />
						</svg>
					</button>
					<span class="page-title">{pageTitle}</span>
				</div>
				<div class="bar-right">
					<NotificationCenter />
					<ThemeToggle />
				</div>
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

<!-- Global, mounted once: action feedback + confirmation dialogs. -->
<Toaster />
<ConfirmHost />

<style>
	.app {
		display: grid;
		grid-template-columns: var(--sidebar-width) 1fr;
		min-height: 100dvh;
		/* Animate the rail collapse/expand (snaps where unsupported). */
		transition: grid-template-columns var(--collapse-dur) var(--ease-out);
	}
	.main {
		display: flex;
		flex-direction: column;
		background-color: var(--color-bg);
		min-width: 0;
	}
	.topbar {
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: var(--space-3);
		padding: var(--space-4) var(--space-6);
		border-bottom: var(--border-width-1) solid var(--color-divider);
		position: sticky;
		top: 0;
		z-index: 50;
		background-color: color-mix(in srgb, var(--color-bg) 88%, transparent);
		backdrop-filter: blur(8px);
	}
	.bar-left,
	.bar-right {
		display: flex;
		align-items: center;
		gap: var(--space-3);
	}
	.page-title {
		font-size: var(--text-base);
		font-weight: var(--weight-semibold);
		letter-spacing: var(--tracking-snug);
		color: var(--color-text);
	}
	.hamburger {
		display: none;
		align-items: center;
		justify-content: center;
		width: 34px;
		height: 34px;
		border: var(--border-width-1) solid var(--color-border);
		border-radius: var(--radius-md);
		background-color: var(--color-surface);
		color: var(--color-text-muted);
		cursor: pointer;
		transition:
			border-color var(--dur-fast) var(--ease-out),
			color var(--dur-fast) var(--ease-out);
	}
	.hamburger:hover {
		border-color: var(--color-border-strong);
		color: var(--color-text);
	}
	.hamburger:focus-visible {
		outline: none;
		box-shadow: var(--shadow-focus);
	}
	@media (max-width: 720px) {
		.hamburger {
			display: inline-flex;
		}
	}
	.content {
		flex: 1;
		padding: var(--space-4) var(--space-6);
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
