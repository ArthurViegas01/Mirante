<script>
	// status: success | warning | danger | info
	let { status = 'info', label = '', children } = $props();
</script>

<span class="badge {status}">
	<span class="dot" aria-hidden="true"></span>
	{label}{@render children?.()}
</span>

<style>
	.badge {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		border-radius: var(--radius-full);
		padding: 2px 9px;
		font-size: 11px;
		font-weight: var(--weight-semibold);
		white-space: nowrap;
	}
	.dot {
		width: 6px;
		height: 6px;
		border-radius: var(--radius-full);
	}
	.success {
		background-color: var(--color-success-bg);
		color: var(--color-success-text);
	}
	.success .dot {
		background-color: var(--color-success);
	}
	.warning {
		background-color: var(--color-warning-bg);
		color: var(--color-warning-text);
	}
	.warning .dot {
		background-color: var(--color-warning);
	}
	.danger {
		background-color: var(--color-danger-bg);
		color: var(--color-danger-text);
	}
	.danger .dot {
		background-color: var(--color-danger);
	}
	.info {
		background-color: var(--color-info-bg);
		color: var(--color-info-text);
	}
	.info .dot {
		background-color: var(--color-info);
	}
</style>
