<script>
	// value: 0..1; variant: success | warning | danger | accent
	let { value = 0, variant = 'accent' } = $props();
	let pct = $derived(Math.max(0, Math.min(1, value)) * 100);
</script>

<div class="track" role="progressbar" aria-valuenow={Math.round(pct)} aria-valuemin="0" aria-valuemax="100">
	<div class="fill {variant}" style="width: {pct}%"></div>
</div>

<style>
	.track {
		height: 5px;
		background-color: var(--color-divider);
		border-radius: var(--radius-full);
		overflow: hidden;
	}
	.fill {
		height: 100%;
		border-radius: var(--radius-full);
		transition: width var(--dur-base) var(--ease-out);
	}
	.accent {
		background-color: var(--color-accent);
	}
	.success {
		background-color: var(--color-success);
	}
	.warning {
		background-color: var(--color-warning);
	}
	.danger {
		background-color: var(--color-danger);
	}
</style>
