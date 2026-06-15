<script>
	import { page } from '$app/stores';
	import { afterNavigate } from '$app/navigation';
	import { monitor } from '$lib/stores/monitor.svelte.js';
	import { session } from '$lib/stores/session.svelte.js';
	import BrandMark from '$lib/components/BrandMark.svelte';
	import UserMenu from '$lib/components/UserMenu.svelte';

	// On mobile the sidebar is an off-canvas drawer controlled by the layout.
	let { open = $bindable(false) } = $props();

	// Close after any navigation (drawer use) and on Escape.
	afterNavigate(() => {
		open = false;
	});
	$effect(() => {
		if (!open) return;
		function onKey(e) {
			if (e.key === 'Escape') open = false;
		}
		document.addEventListener('keydown', onKey);
		return () => document.removeEventListener('keydown', onKey);
	});

	const items = [
		{
			href: '/',
			label: 'Início',
			icon: 'm3 9 9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z M9 22V12h6v10'
		},
		{
			href: '/projetos',
			label: 'Projetos',
			icon: 'M3 7a2 2 0 0 1 2-2h4l2 2h6a2 2 0 0 1 2 2v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z'
		},
		{
			href: '/tarefas',
			label: 'Tarefas',
			icon: 'M9 11l3 3L22 4M21 12v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11'
		},
		{
			href: '/custos',
			label: 'Custos',
			icon: 'M12 1v22M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6'
		},
		{
			href: '/vagas',
			label: 'Vagas',
			icon: 'M3 7h18v13H3zM8 7V5a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2'
		},
		{
			href: '/cv',
			label: 'CV',
			icon: 'M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8zM14 2v6h6M16 13H8M16 17H8'
		},
		{
			href: '/candidaturas',
			label: 'Candidaturas',
			icon: 'M3 5h18l-7 8v5l-4 2v-7z'
		}
	];

	function active(path, href) {
		return path === href || path.startsWith(href + '/');
	}
</script>

<div class="scrim" class:show={open} onclick={() => (open = false)} aria-hidden="true"></div>

<nav class="sidebar" class:open aria-label="Navegação principal">
	<div class="brand">
		<BrandMark tag="MIRANTE" />
		<span
			class="live"
			class:on={monitor.connected}
			title={monitor.connected ? 'Ao vivo' : 'Offline'}
			aria-label="Indicador ao vivo"
		></span>
		<button class="close-drawer" onclick={() => (open = false)} aria-label="Fechar menu">
			<svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
				<path d="M18 6 6 18M6 6l12 12" />
			</svg>
		</button>
	</div>

	<ul class="nav">
		{#each items as item (item.href)}
			<li>
				<a
					href={item.href}
					class:active={active($page.url.pathname, item.href)}
					aria-current={active($page.url.pathname, item.href) ? 'page' : undefined}
				>
					<svg
						viewBox="0 0 24 24"
						width="16"
						height="16"
						fill="none"
						stroke="currentColor"
						stroke-width="1.75"
						stroke-linecap="round"
						stroke-linejoin="round"
						aria-hidden="true"
					>
						<path d={item.icon} />
					</svg>
					<span>{item.label}</span>
				</a>
			</li>
		{/each}
		{#if session.isAdmin}
			<li>
				<a
					href="/admin/usuarios"
					class:active={active($page.url.pathname, '/admin/usuarios')}
					aria-current={active($page.url.pathname, '/admin/usuarios') ? 'page' : undefined}
				>
					<svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
						<path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2M9 11a4 4 0 1 0 0-8 4 4 0 0 0 0 8M23 21v-2a4 4 0 0 0-3-3.87M16 3.13a4 4 0 0 1 0 7.75" />
					</svg>
					<span>Usuários</span>
				</a>
			</li>
		{/if}
	</ul>

	<div class="footer">
		<UserMenu />
	</div>
</nav>

<style>
	/* The sidebar is always dark (Ink-950) per the design system, regardless of theme. */
	.sidebar {
		display: flex;
		flex-direction: column;
		height: 100dvh;
		position: sticky;
		top: 0;
		background-color: var(--ink-950);
		border-right: var(--border-width-1) solid rgba(255, 255, 255, 0.05);
		padding: 18px 14px;
	}

	.brand {
		display: flex;
		align-items: center;
		gap: 7px;
		padding: 4px 6px 16px;
		color: #ffffff;
		--mark-fill: var(--glow);
		--word-size: 16px;
		--tag-color: rgba(255, 255, 255, 0.4);
	}
	.live {
		width: 7px;
		height: 7px;
		margin-left: auto;
		border-radius: var(--radius-full);
		background-color: rgba(255, 255, 255, 0.25);
	}
	.live.on {
		background-color: var(--color-live);
		box-shadow: 0 0 0 4px var(--color-live-halo);
		animation: live-pulse 2.4s var(--ease-in-out) infinite;
	}
	@keyframes live-pulse {
		0%,
		100% {
			box-shadow: 0 0 0 3px var(--color-live-halo);
		}
		50% {
			box-shadow: 0 0 0 7px color-mix(in srgb, var(--glow) 8%, transparent);
		}
	}

	.nav {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: 4px;
		flex: 1;
	}
	.nav a {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 7px 9px;
		border-radius: 7px;
		font-size: 13.5px;
		color: rgba(255, 255, 255, 0.72);
		text-decoration: none;
		transition:
			background-color var(--dur-fast) var(--ease-out),
			color var(--dur-fast) var(--ease-out);
	}
	.nav a svg {
		color: rgba(255, 255, 255, 0.5);
	}
	.nav a:hover {
		background-color: rgba(255, 255, 255, 0.04);
		color: #ffffff;
	}
	.nav a:hover svg {
		color: #ffffff;
	}
	.nav a.active {
		background-color: rgba(94, 234, 212, 0.1);
		color: var(--glow);
	}
	.nav a.active svg {
		color: var(--glow);
	}
	.nav a:focus-visible {
		outline: none;
		box-shadow: var(--shadow-focus);
	}

	.footer {
		border-top: var(--border-width-1) solid rgba(255, 255, 255, 0.06);
		padding-top: 10px;
		margin-top: 12px;
	}

	/* Mobile drawer affordances (hidden on desktop). */
	.close-drawer {
		display: none;
		align-items: center;
		justify-content: center;
		width: 32px;
		height: 32px;
		margin-left: 4px;
		border: none;
		border-radius: var(--radius-md);
		background: transparent;
		color: rgba(255, 255, 255, 0.6);
		cursor: pointer;
	}
	.close-drawer:hover {
		color: #ffffff;
		background-color: rgba(255, 255, 255, 0.06);
	}
	.close-drawer:focus-visible {
		outline: none;
		box-shadow: var(--shadow-focus);
	}
	.scrim {
		display: none;
	}

	@media (max-width: 720px) {
		.sidebar {
			position: fixed;
			top: 0;
			left: 0;
			bottom: 0;
			width: min(82vw, 300px);
			z-index: 120;
			transform: translateX(-100%);
			transition: transform var(--dur-base) var(--ease-out);
			border-right: var(--border-width-1) solid rgba(255, 255, 255, 0.08);
		}
		.sidebar.open {
			transform: translateX(0);
			box-shadow: var(--shadow-xl);
		}
		.scrim {
			display: block;
			position: fixed;
			inset: 0;
			z-index: 110;
			background-color: color-mix(in srgb, var(--ink-950) 55%, transparent);
			opacity: 0;
			pointer-events: none;
			transition: opacity var(--dur-base) var(--ease-out);
		}
		.scrim.show {
			opacity: 1;
			pointer-events: auto;
		}
		.close-drawer {
			display: inline-flex;
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.sidebar,
		.scrim {
			transition: none;
		}
	}
</style>
