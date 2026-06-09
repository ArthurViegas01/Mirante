<script>
	import { page } from '$app/stores';
	import { monitor } from '$lib/stores/monitor.svelte.js';

	const items = [
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
		}
	];

	function active(path, href) {
		return path === href || path.startsWith(href + '/');
	}
</script>

<nav class="sidebar" aria-label="Navegação principal">
	<div class="brand">
		<svg class="mark" viewBox="0 0 24 24" width="22" height="22" aria-hidden="true">
			<path
				d="M12 1.5 14.2 9.8 22.5 12 14.2 14.2 12 22.5 9.8 14.2 1.5 12 9.8 9.8Z"
				fill="var(--glow)"
			/>
		</svg>
		<span class="wordmark">mirante</span>
		<span class="tag">MIRANTE</span>
		<span
			class="live"
			class:on={monitor.connected}
			title={monitor.connected ? 'Ao vivo' : 'Offline'}
			aria-label="Indicador ao vivo"
		></span>
	</div>

	<ul class="nav">
		{#each items as item (item.href)}
			<li>
				<a href={item.href} class:active={active($page.url.pathname, item.href)}>
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
	</ul>

	<div class="footer">
		<span class="org">by Lumni</span>
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
	}
	.wordmark {
		font-family: var(--font-serif);
		font-style: italic;
		font-size: 16px;
		font-weight: 400;
		color: #ffffff;
		letter-spacing: -0.01em;
	}
	.tag {
		font-family: var(--font-mono);
		font-size: 9.5px;
		letter-spacing: var(--tracking-eyebrow);
		color: rgba(255, 255, 255, 0.4);
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
		padding-top: 12px;
		margin-top: 12px;
	}
	.org {
		font-family: var(--font-mono);
		font-size: 11px;
		color: rgba(255, 255, 255, 0.4);
	}
</style>
