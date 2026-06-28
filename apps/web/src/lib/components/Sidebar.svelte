<script>
	import { page } from '$app/stores';
	import { afterNavigate } from '$app/navigation';
	import { onMount } from 'svelte';
	import { session } from '$lib/stores/session.svelte.js';
	import BrandMark from '$lib/components/BrandMark.svelte';
	import UserMenu from '$lib/components/UserMenu.svelte';

	// On mobile the sidebar is an off-canvas drawer controlled by the layout.
	let { open = $bindable(false) } = $props();

	// Desktop-only "icon rail" collapse. The visual state is driven entirely by
	// the `data-sidebar` attribute on <html> (set pre-paint by the inline script
	// in app.html, so there is no flash and no hydration mismatch); this flag only
	// mirrors it for the toggle button's accessible name/state.
	let collapsed = $state(false);
	onMount(() => {
		collapsed = document.documentElement.dataset.sidebar === 'collapsed';
	});
	function toggleCollapse() {
		collapsed = !collapsed;
		const root = document.documentElement;
		if (collapsed) root.dataset.sidebar = 'collapsed';
		else delete root.dataset.sidebar;
		try {
			localStorage.setItem('sidebar-collapsed', collapsed ? '1' : '0');
		} catch (e) {
			/* ignore */
		}
	}

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

	// Two visible sections of primary nav.
	const projetos = [
		{
			href: '/',
			label: '/Início',
			icon: 'm3 9 9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z M9 22V12h6v10'
		},
		{
			href: '/projetos',
			label: '/Projetos',
			icon: 'M3 7a2 2 0 0 1 2-2h4l2 2h6a2 2 0 0 1 2 2v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z'
		},
		{
			href: '/tarefas',
			label: '/Tarefas',
			icon: 'M9 11l3 3L22 4M21 12v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11'
		},
		{
			href: '/custos',
			label: '/Custos',
			icon: 'M12 1v22M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6'
		}
	];
	const perfil = [
		{
			href: '/perfil',
			label: '/Meu Perfil',
			icon: 'M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2M16 7a4 4 0 1 1-8 0 4 4 0 0 1 8 0'
		},
		{
			href: '/vagas',
			label: '/Vagas',
			icon: 'M3 7h18v13H3zM8 7V5a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2'
		},
		{
			href: '/cv',
			label: '/Currículo',
			icon: 'M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8zM14 2v6h6M16 13H8M16 17H8'
		},
		{
			href: '/candidaturas',
			label: '/Candidaturas',
			icon: 'M3 5h18l-7 8v5l-4 2v-7z'
		}
	];
	// Owner-only section (rendered only when session.isAdmin).
	const administracao = [
		{
			href: '/admin/usuarios',
			label: '/Usuários',
			icon: 'M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2M9 11a4 4 0 1 0 0-8 4 4 0 0 0 0 8M23 21v-2a4 4 0 0 0-3-3.87M16 3.13a4 4 0 0 1 0 7.75'
		}
	];

	function active(path, href) {
		return path === href || path.startsWith(href + '/');
	}
</script>

<div class="scrim" class:show={open} onclick={() => (open = false)} aria-hidden="true"></div>

<nav class="sidebar" class:open aria-label="Navegação principal">
	<div class="brand">
		<BrandMark tag="by lumni" pulse />
		<button
			class="rail-toggle"
			onclick={toggleCollapse}
			aria-label={collapsed ? 'Expandir menu' : 'Recolher menu'}
			aria-expanded={!collapsed}
			title={collapsed ? 'Expandir' : 'Recolher'}
		>
			<svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
				<path d="m15 18-6-6 6-6" />
			</svg>
		</button>
		<button class="close-drawer" onclick={() => (open = false)} aria-label="Fechar menu">
			<svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
				<path d="M18 6 6 18M6 6l12 12" />
			</svg>
		</button>
	</div>

	<!-- Áreas de navegação: Projetos e Perfil. -->
	<div class="nav-groups">
		<p class="section-label">Projetos</p>
		<ul class="nav">
			{#each projetos as item (item.href)}{@render navItem(item)}{/each}
		</ul>
		<p class="section-label">Perfil</p>
		<ul class="nav">
			{#each perfil as item (item.href)}{@render navItem(item)}{/each}
		</ul>
	</div>

	<!-- Divider between the nav above and the admin/account area below. -->
	<div class="nav-sep" aria-hidden="true"></div>

	<!-- Owner-only área: section label + links visible only to the admin. -->
	{#if session.isAdmin}
		<p class="section-label">Administração</p>
		<ul class="nav nav-bottom">
			{#each administracao as item (item.href)}{@render navItem(item)}{/each}
		</ul>
	{/if}

	<div class="footer">
		<UserMenu />
	</div>
</nav>

{#snippet navItem(item)}
	<li>
		<a
			href={item.href}
			data-label={item.label}
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
{/snippet}

<style>
	/* The sidebar is always dark (Ink-950) per the design system, regardless of theme. */
	.sidebar {
		display: flex;
		flex-direction: column;
		height: 100dvh;
		position: sticky;
		top: 0;
		/* Above the content column so collapsed-mode tooltips/flyouts overlay it. */
		z-index: 100;
		background-color: var(--ink-950);
		border-right: var(--border-width-1) solid rgba(255, 255, 255, 0.05);
		padding: 18px 14px 10px;
	}

	.brand {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 7px;
		/* Height tuned so the divider lines up with the topbar's bottom border:
		   sidebar padding-top (18) + brand height (49) = topbar height (66 + 1px border). */
		height: 49px;
		/* Negative side margins + matching padding extend the bottom border edge
		   to edge of the rail (the sidebar pads 14px on each side). */
		margin: 0 -14px 12px;
		padding: 0 20px;
		border-bottom: var(--border-width-1) solid rgba(255, 255, 255, 0.09);
		color: #ffffff;
		--mark-fill: var(--glow);
		--word-size: 20px;
		--tag-color: rgba(255, 255, 255, 0.4);
	}
	/* Collapse/expand control. Chevron points left to collapse; CSS rotates it to
	   point right when the rail is collapsed. */
	.rail-toggle {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 26px;
		height: 26px;
		flex-shrink: 0;
		border: var(--border-width-1) solid rgba(255, 255, 255, 0.12);
		border-radius: var(--radius-full);
		background-color: rgba(255, 255, 255, 0.05);
		color: rgba(255, 255, 255, 0.65);
		cursor: pointer;
		transition:
			background-color var(--dur-fast) var(--ease-out),
			border-color var(--dur-fast) var(--ease-out),
			color var(--dur-fast) var(--ease-out);
	}
	.rail-toggle:hover {
		background-color: rgba(255, 255, 255, 0.1);
		border-color: rgba(255, 255, 255, 0.2);
		color: #ffffff;
	}
	.rail-toggle:focus-visible {
		outline: none;
		box-shadow: var(--shadow-focus);
	}
	.rail-toggle svg {
		transition: transform var(--collapse-dur) var(--ease-out);
	}

	/* Section kicker for each area (mono caps), inset to align with item labels. */
	.section-label {
		margin: 2px 0 6px;
		padding: 0 9px;
		font-family: var(--font-mono);
		font-size: var(--text-xs);
		letter-spacing: var(--tracking-eyebrow);
		text-transform: uppercase;
		color: rgba(255, 255, 255, 0.5);
		white-space: nowrap;
		overflow: hidden;
		transition: opacity var(--collapse-dur) var(--ease-out);
	}
	/* Extra breathing room when a section label follows another section's list. */
	.nav + .section-label {
		margin-top: 16px;
	}

	/* Holds the projects area; flex:1 pushes the profile area down to the bottom. */
	.nav-groups {
		flex: 1;
		display: flex;
		flex-direction: column;
	}
	.nav {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: 4px;
	}
	/* Divider between the projects area and the profile area. */
	.nav-sep {
		height: var(--border-width-1);
		background-color: rgba(255, 255, 255, 0.09);
		margin: 8px 9px;
	}
	/* Pinned to the bottom of the rail, above the user footer. */
	.nav-bottom {
		flex: 0 0 auto;
	}
	.nav a {
		position: relative;
		display: flex;
		align-items: center;
		padding: 7px 9px;
		border-radius: 7px;
		font-size: 13.5px;
		color: rgba(255, 255, 255, 0.72);
		text-decoration: none;
		transition:
			background-color var(--dur-fast) var(--ease-out),
			color var(--dur-fast) var(--ease-out),
			padding var(--collapse-dur) var(--ease-out);
	}
	.nav a svg {
		flex-shrink: 0;
		color: rgba(255, 255, 255, 0.5);
	}
	/* The label is a flex item whose width + opacity animate, so it slides and
	   fades out as the rail collapses (the icon re-centers via the padding above). */
	.nav a span {
		margin-left: 10px;
		max-width: 130px;
		overflow: hidden;
		white-space: nowrap;
		transition:
			margin-left var(--collapse-dur) var(--ease-out),
			max-width var(--collapse-dur) var(--ease-out),
			opacity var(--collapse-dur) var(--ease-out);
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

	/* No border here: the .nav-sep above the profile area is the divider now. */
	.footer {
		padding-top: 2px;
	}

	/* ============================================================
	   COLLAPSED — desktop icon rail. Driven by the <html> attribute
	   so it is correct on first paint. Mobile keeps the full drawer.
	   ============================================================ */
	@media (min-width: 721px) {
		/* Brand stacks: mark on top, toggle below, both centered. */
		:global(html[data-sidebar='collapsed']) .brand {
			flex-direction: column;
			align-items: center;
			justify-content: center;
			height: auto;
			gap: 12px;
			padding: 0 0 14px;
		}
		/* Hide the whole word/tag stack (not just its text) so the lockup is just
		   the mark and centers exactly in the rail. */
		:global(html[data-sidebar='collapsed'] .brand .word-stack) {
			display: none;
		}
		/* Fade the labels but keep their box, so the empty space turns into a
		   subtle separator between the icon groups. */
		:global(html[data-sidebar='collapsed']) .section-label {
			opacity: 0;
		}
		/* The admin label sits right after the divider, so its (invisible) box
		   would leave a big gap above the Usuários icon. The divider already marks
		   the boundary, so drop this label's box entirely. */
		:global(html[data-sidebar='collapsed']) .nav-sep + .section-label {
			display: none;
		}
		/* Icons only: the label collapses to zero width + fades (see base rule),
		   and the growing side padding re-centers the icon. The text stays in the
		   DOM (not display:none), so the link keeps its accessible name. */
		:global(html[data-sidebar='collapsed']) .nav a {
			padding: 9px 14px;
		}
		:global(html[data-sidebar='collapsed']) .nav a span {
			margin-left: 0;
			max-width: 0;
			opacity: 0;
		}
		/* Hover/focus tooltip bubble carrying the label, to the right of the icon. */
		:global(html[data-sidebar='collapsed']) .nav a::after {
			content: attr(data-label);
			position: absolute;
			left: calc(100% + 12px);
			top: 50%;
			transform: translateY(-50%);
			padding: 5px 9px;
			border-radius: var(--radius-sm);
			background-color: var(--ink-900);
			color: #ffffff;
			font-size: 12px;
			font-weight: var(--weight-medium);
			line-height: 1;
			white-space: nowrap;
			box-shadow: var(--shadow-lg);
			opacity: 0;
			pointer-events: none;
			transition: opacity var(--dur-fast) var(--ease-out);
			z-index: 200;
		}
		:global(html[data-sidebar='collapsed']) .nav a::before {
			content: '';
			position: absolute;
			left: calc(100% + 7px);
			top: 50%;
			transform: translateY(-50%) rotate(45deg);
			width: 8px;
			height: 8px;
			background-color: var(--ink-900);
			opacity: 0;
			pointer-events: none;
			transition: opacity var(--dur-fast) var(--ease-out);
			z-index: 200;
		}
		:global(html[data-sidebar='collapsed']) .nav a:hover::after,
		:global(html[data-sidebar='collapsed']) .nav a:hover::before,
		:global(html[data-sidebar='collapsed']) .nav a:focus-visible::after,
		:global(html[data-sidebar='collapsed']) .nav a:focus-visible::before {
			opacity: 1;
		}
		/* Flip the chevron to point right (expand). */
		:global(html[data-sidebar='collapsed']) .rail-toggle svg {
			transform: rotate(180deg);
		}
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
		/* The rail-collapse control is desktop-only; the drawer has its own close. */
		.rail-toggle {
			display: none;
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
