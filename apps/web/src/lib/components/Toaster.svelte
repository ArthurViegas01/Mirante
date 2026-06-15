<script>
	import { toasts } from '$lib/stores/toast.svelte.js';
	import { fly, fade } from 'svelte/transition';

	const icons = {
		success: 'M20 6 9 17l-5-5',
		error: 'M18 6 6 18M6 6l12 12',
		info: 'M12 16v-4M12 8h.01M12 22a10 10 0 1 0 0-20 10 10 0 0 0 0 20z'
	};
</script>

<div class="toaster" aria-live="polite" aria-atomic="false">
	{#each toasts.items as t (t.id)}
		<div
			class="toast {t.type}"
			role={t.type === 'error' ? 'alert' : 'status'}
			in:fly={{ y: 12, duration: 240 }}
			out:fade={{ duration: 160 }}
		>
			<svg class="ic" viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
				<path d={icons[t.type] ?? icons.info} />
			</svg>
			<span class="msg">{t.message}</span>
			<button class="x" onclick={() => toasts.dismiss(t.id)} aria-label="Dispensar">
				<svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
					<path d="M18 6 6 18M6 6l12 12" />
				</svg>
			</button>
		</div>
	{/each}
</div>

<style>
	.toaster {
		position: fixed;
		bottom: var(--space-5);
		right: var(--space-5);
		z-index: 200;
		display: flex;
		flex-direction: column;
		gap: var(--space-2);
		max-width: min(380px, calc(100vw - 2 * var(--space-5)));
		pointer-events: none;
	}
	.toast {
		pointer-events: auto;
		display: flex;
		align-items: flex-start;
		gap: var(--space-3);
		padding: 12px 12px 12px 14px;
		background-color: var(--color-surface-elevated);
		border: var(--border-width-1) solid var(--color-border);
		border-left-width: 3px;
		border-radius: var(--radius-md);
		box-shadow: var(--shadow-lg);
	}
	.toast.success {
		border-left-color: var(--color-success);
	}
	.toast.error {
		border-left-color: var(--color-danger);
	}
	.toast.info {
		border-left-color: var(--color-info);
	}
	.ic {
		flex-shrink: 0;
		margin-top: 1px;
	}
	.toast.success .ic {
		color: var(--color-success);
	}
	.toast.error .ic {
		color: var(--color-danger);
	}
	.toast.info .ic {
		color: var(--color-info);
	}
	.msg {
		flex: 1;
		font-size: var(--text-sm);
		line-height: var(--leading-normal);
		color: var(--color-text);
	}
	.x {
		flex-shrink: 0;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 22px;
		height: 22px;
		margin: -2px -2px 0 0;
		border: none;
		border-radius: var(--radius-sm);
		background: transparent;
		color: var(--color-text-muted);
		cursor: pointer;
		transition: color var(--dur-fast) var(--ease-out);
	}
	.x:hover {
		color: var(--color-text);
	}
	.x:focus-visible {
		outline: none;
		box-shadow: var(--shadow-focus);
	}
</style>
