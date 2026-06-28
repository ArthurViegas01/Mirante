<script>
	import { goto } from '$app/navigation';
	import { api, setCsrf } from '$lib/api.js';
	import { session } from '$lib/stores/session.svelte.js';

	let open = $state(false);
	let busy = $state(false);
	let root;

	function onDocPointer(e) {
		if (root && !root.contains(e.target)) open = false;
	}
	function onKey(e) {
		if (e.key === 'Escape') open = false;
	}

	// Listeners live only while the menu is open. Capture phase so a click
	// anywhere outside closes it; clicks on the trigger stay inside `root`.
	$effect(() => {
		if (!open) return;
		document.addEventListener('pointerdown', onDocPointer, true);
		document.addEventListener('keydown', onKey);
		return () => {
			document.removeEventListener('pointerdown', onDocPointer, true);
			document.removeEventListener('keydown', onKey);
		};
	});

	async function logout() {
		busy = true;
		try {
			await api('/api/auth/logout', { method: 'POST' });
		} catch (e) {
			// Clear locally even if the request fails; the cookie may already be gone.
		}
		setCsrf('');
		session.clear();
		open = false;
		busy = false;
		goto('/login');
	}
</script>

<div class="um" bind:this={root}>
	<button
		class="trigger"
		onclick={() => (open = !open)}
		aria-haspopup="menu"
		aria-expanded={open}
	>
		<span class="avatar" aria-hidden="true">{session.initials}</span>
		<span class="who">
			<span class="name">{session.displayName}</span>
		</span>
		<svg class="chev" class:up={open} viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
			<path d="m6 9 6 6 6-6" />
		</svg>
	</button>

	{#if open}
		<div class="menu" role="menu">
			<div class="head">
				<span class="avatar lg" aria-hidden="true">{session.initials}</span>
				<div class="head-text">
					<span class="name">{session.displayName}</span>
					{#if session.user?.email}<span class="email">{session.user.email}</span>{/if}
				</div>
			</div>
			<div class="sep"></div>
			<button class="item danger" role="menuitem" onclick={logout} disabled={busy}>
				<svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
					<path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" />
					<path d="m16 17 5-5-5-5" />
					<path d="M21 12H9" />
				</svg>
				<span>{busy ? 'Saindo…' : 'Sair'}</span>
			</button>
		</div>
	{/if}
</div>

<style>
	.um {
		position: relative;
	}
	.trigger {
		display: flex;
		align-items: center;
		gap: 9px;
		width: 100%;
		padding: 8px;
		border: none;
		border-radius: 9px;
		background: transparent;
		text-align: left;
		cursor: pointer;
		color: #ffffff;
		transition: background-color var(--dur-fast) var(--ease-out);
	}
	.trigger:hover {
		background-color: rgba(255, 255, 255, 0.05);
	}
	.trigger:focus-visible {
		outline: none;
		box-shadow: var(--shadow-focus);
	}
	.avatar {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 28px;
		height: 28px;
		flex-shrink: 0;
		border-radius: var(--radius-full);
		background: linear-gradient(135deg, var(--beam-600), var(--glow));
		color: var(--ink-950);
		font-family: var(--font-mono);
		font-size: 11px;
		font-weight: var(--weight-semibold);
		letter-spacing: 0;
	}
	.avatar.lg {
		width: 34px;
		height: 34px;
		font-size: 12.5px;
	}
	.who {
		display: flex;
		flex-direction: column;
		gap: 1px;
		min-width: 0;
		flex: 1;
	}
	.who .name {
		font-size: 12.5px;
		font-weight: var(--weight-medium);
		color: #ffffff;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.chev {
		color: rgba(255, 255, 255, 0.5);
		flex-shrink: 0;
		transition: transform var(--dur-fast) var(--ease-out);
	}
	.chev.up {
		transform: rotate(180deg);
	}

	/* Collapsed rail (desktop): avatar only, and the dropdown opens to the side
	   as a flyout instead of upward. Driven by the <html> attribute. */
	@media (min-width: 721px) {
		:global(html[data-sidebar='collapsed']) .trigger {
			justify-content: center;
			padding: 8px 0;
		}
		/* Keep the name as the trigger's accessible name (avatar + chevron are
		   aria-hidden), so hide it visually rather than with display:none. */
		:global(html[data-sidebar='collapsed']) .who {
			position: absolute;
			width: 1px;
			height: 1px;
			padding: 0;
			margin: -1px;
			overflow: hidden;
			clip: rect(0, 0, 0, 0);
			white-space: nowrap;
			border: 0;
		}
		:global(html[data-sidebar='collapsed']) .chev {
			display: none;
		}
		:global(html[data-sidebar='collapsed']) .menu {
			left: calc(100% + 12px);
			right: auto;
			bottom: 0;
			min-width: 220px;
		}
	}

	.menu {
		position: absolute;
		bottom: calc(100% + 8px);
		left: 0;
		right: 0;
		min-width: 200px;
		background-color: var(--color-surface-elevated);
		border: var(--border-width-1) solid var(--color-border);
		border-radius: var(--radius-lg);
		box-shadow: var(--shadow-lg);
		padding: 6px;
		z-index: 40;
		animation: menu-in var(--dur-fast) var(--ease-out);
	}
	@keyframes menu-in {
		from {
			opacity: 0;
			transform: translateY(6px) scale(0.98);
		}
		to {
			opacity: 1;
			transform: translateY(0) scale(1);
		}
	}
	.head {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 8px 8px 10px;
	}
	.head-text {
		display: flex;
		flex-direction: column;
		gap: 1px;
		min-width: 0;
	}
	.head-text .name {
		font-size: 13px;
		font-weight: var(--weight-semibold);
		color: var(--color-text);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.head-text .email {
		font-family: var(--font-mono);
		font-size: 11px;
		color: var(--color-text-muted);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.sep {
		height: 1px;
		background-color: var(--color-divider);
		margin: 0 2px 6px;
	}
	.item {
		display: flex;
		align-items: center;
		gap: 9px;
		width: 100%;
		padding: 8px 8px;
		border: none;
		border-radius: var(--radius-sm);
		background: transparent;
		color: var(--color-text-secondary);
		font-family: var(--font-sans);
		font-size: 13px;
		cursor: pointer;
		transition:
			background-color var(--dur-fast) var(--ease-out),
			color var(--dur-fast) var(--ease-out);
	}
	.item:hover:not(:disabled) {
		background-color: var(--color-surface-sunken);
		color: var(--color-text);
	}
	.item.danger:hover:not(:disabled) {
		background-color: var(--color-danger-bg);
		color: var(--color-danger-text);
	}
	.item:focus-visible {
		outline: none;
		box-shadow: var(--shadow-focus);
	}
	.item:disabled {
		opacity: 0.6;
		cursor: progress;
	}
</style>
