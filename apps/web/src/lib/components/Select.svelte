<script>
	let { value = $bindable(''), label = '', options = [] } = $props();
</script>

<label class="field">
	{#if label}<span class="label">{label}</span>{/if}
	<select bind:value>
		{#each options as o (o.value)}
			<option value={o.value}>{o.label}</option>
		{/each}
	</select>
</label>

<style>
	.field {
		display: flex;
		flex-direction: column;
		gap: var(--space-1);
	}
	.label {
		font-size: 11.5px;
		font-weight: var(--weight-medium);
		color: var(--color-text-secondary);
	}
	select {
		font-family: var(--font-sans);
		font-size: 13px;
		padding: 8px 10px;
		background-color: var(--color-surface);
		color: var(--color-text);
		border: var(--border-width-1) solid var(--color-border);
		border-radius: var(--radius-md);
		cursor: pointer;
	}
	select:hover {
		border-color: var(--color-border-strong);
	}
	select:focus {
		outline: none;
		border-color: var(--color-primary);
		box-shadow: var(--shadow-focus);
	}
</style>
