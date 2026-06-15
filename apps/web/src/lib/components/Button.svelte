<script>
	let {
		variant = 'primary',
		size = 'md',
		type = 'button',
		disabled = false,
		full = false,
		onclick,
		children
	} = $props();
</script>

<button class="btn {variant} {size}" class:full {type} {disabled} {onclick}>
	{@render children?.()}
</button>

<style>
	.btn {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		gap: 8px;
		line-height: 1.2;
		white-space: nowrap;
		font-family: var(--font-sans);
		font-weight: var(--weight-semibold);
		letter-spacing: -0.005em;
		border: var(--border-width-1) solid transparent;
		border-radius: var(--radius-md);
		cursor: pointer;
		transition:
			background-color var(--dur-instant) var(--ease-out),
			border-color var(--dur-fast) var(--ease-out),
			transform var(--dur-instant) var(--ease-out);
	}
	.btn:focus-visible {
		outline: none;
		box-shadow: var(--shadow-focus);
	}
	.btn:active:not(:disabled) {
		transform: scale(0.98);
	}
	.btn:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}
	.btn.full {
		width: 100%;
	}

	.sm {
		padding: 8px 14px;
		font-size: 13px;
	}
	.md {
		padding: 10px 18px;
		font-size: 14px;
	}
	.lg {
		padding: 14px 24px;
		font-size: 15px;
	}

	.primary {
		background-color: var(--color-primary);
		color: var(--color-primary-text);
	}
	.primary:hover:not(:disabled) {
		background-color: var(--color-primary-hover);
	}
	.secondary {
		background-color: var(--color-surface);
		border-color: var(--color-border);
		color: var(--color-text);
	}
	.secondary:hover:not(:disabled) {
		border-color: var(--color-border-strong);
		background-color: var(--color-surface-sunken);
	}
	.ghost {
		background-color: transparent;
		color: var(--color-text-secondary);
	}
	.ghost:hover:not(:disabled) {
		background-color: var(--color-surface-sunken);
		color: var(--color-text);
	}
	.danger {
		/* primary-text flips white (light) / ink-950 (dark) for legibility on the
		   theme-shifted danger color. */
		background-color: var(--color-danger);
		color: var(--color-primary-text);
	}
	.danger:hover:not(:disabled) {
		background-color: color-mix(in srgb, var(--color-danger) 85%, var(--ink-950));
	}
</style>
