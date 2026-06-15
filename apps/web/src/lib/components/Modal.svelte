<script>
	import { fade, scale } from 'svelte/transition';

	let {
		open = $bindable(false),
		title = '',
		size = 'md',
		onclose,
		children,
		footer
	} = $props();

	let dialogEl = $state(null);

	function close() {
		open = false;
		onclose?.();
	}

	function focusables() {
		if (!dialogEl) return [];
		const sel =
			'a[href], button:not([disabled]), input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])';
		return [...dialogEl.querySelectorAll(sel)].filter((el) => el.offsetParent !== null);
	}

	function onKey(e) {
		if (e.key === 'Escape') {
			e.stopPropagation();
			close();
		} else if (e.key === 'Tab') {
			const f = focusables();
			if (f.length === 0) return;
			const first = f[0];
			const last = f[f.length - 1];
			if (e.shiftKey && document.activeElement === first) {
				e.preventDefault();
				last.focus();
			} else if (!e.shiftKey && document.activeElement === last) {
				e.preventDefault();
				first.focus();
			}
		}
	}

	// While open: trap focus, lock body scroll, and restore focus on close.
	$effect(() => {
		if (!open) return;
		const prev = document.activeElement;
		const overflow = document.body.style.overflow;
		document.body.style.overflow = 'hidden';
		document.addEventListener('keydown', onKey, true);
		queueMicrotask(() => {
			const f = focusables();
			(f[0] ?? dialogEl)?.focus();
		});
		return () => {
			document.removeEventListener('keydown', onKey, true);
			document.body.style.overflow = overflow;
			prev?.focus?.();
		};
	});
</script>

{#if open}
	<div class="backdrop" transition:fade={{ duration: 160 }} onclick={close} role="presentation">
		<div
			class="dialog size-{size}"
			bind:this={dialogEl}
			transition:scale={{ start: 0.96, duration: 240, opacity: 0 }}
			role="dialog"
			aria-modal="true"
			aria-label={title || undefined}
			tabindex="-1"
			onclick={(e) => e.stopPropagation()}
		>
			{#if title}
				<header class="m-head">
					<h2>{title}</h2>
					<button class="x" onclick={close} aria-label="Fechar">
						<svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
							<path d="M18 6 6 18M6 6l12 12" />
						</svg>
					</button>
				</header>
			{/if}
			<div class="m-body">{@render children?.()}</div>
			{#if footer}<footer class="m-foot">{@render footer()}</footer>{/if}
		</div>
	</div>
{/if}

<style>
	.backdrop {
		position: fixed;
		inset: 0;
		z-index: 150;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: var(--space-5);
		background-color: color-mix(in srgb, var(--ink-950) 55%, transparent);
		backdrop-filter: blur(3px);
	}
	.dialog {
		width: 100%;
		max-height: calc(100dvh - 2 * var(--space-8));
		overflow-y: auto;
		display: flex;
		flex-direction: column;
		background-color: var(--color-surface-elevated);
		border: var(--border-width-1) solid var(--color-border);
		border-radius: var(--radius-2xl);
		box-shadow: var(--shadow-xl);
	}
	.size-sm {
		max-width: 400px;
	}
	.size-md {
		max-width: 540px;
	}
	.size-lg {
		max-width: 720px;
	}
	.m-head {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: var(--space-4);
		padding: var(--space-5) var(--space-6);
		border-bottom: var(--border-width-1) solid var(--color-divider);
	}
	.m-head h2 {
		margin: 0;
		font-size: var(--text-lg);
		font-weight: var(--weight-medium);
		letter-spacing: var(--tracking-snug);
		color: var(--color-text);
	}
	.x {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 32px;
		height: 32px;
		flex-shrink: 0;
		border: none;
		border-radius: var(--radius-md);
		background: transparent;
		color: var(--color-text-muted);
		cursor: pointer;
		transition:
			background-color var(--dur-fast) var(--ease-out),
			color var(--dur-fast) var(--ease-out);
	}
	.x:hover {
		background-color: var(--color-surface-sunken);
		color: var(--color-text);
	}
	.x:focus-visible {
		outline: none;
		box-shadow: var(--shadow-focus);
	}
	.m-body {
		padding: var(--space-6);
	}
	.m-foot {
		display: flex;
		justify-content: flex-end;
		gap: var(--space-3);
		padding: var(--space-4) var(--space-6);
		border-top: var(--border-width-1) solid var(--color-divider);
	}
</style>
