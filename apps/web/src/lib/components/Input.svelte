<script>
	// `type` is intentionally applied one-way (not bound) so the input can carry a
	// dynamic type while still supporting two-way `value` via $bindable.
	let {
		value = $bindable(''),
		type = 'text',
		label = '',
		placeholder = '',
		id = '',
		required = false
	} = $props();
</script>

<label class="field">
	{#if label}<span class="label">{label}</span>{/if}
	<input
		{id}
		{type}
		{placeholder}
		{required}
		{value}
		oninput={(e) => (value = e.currentTarget.value)}
	/>
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
	input {
		font-family: var(--font-sans);
		font-size: 13px;
		padding: 8px 10px;
		background-color: var(--color-surface);
		color: var(--color-text);
		border: var(--border-width-1) solid var(--color-border);
		border-radius: var(--radius-md);
		transition:
			border-color var(--dur-fast) var(--ease-out),
			box-shadow var(--dur-fast) var(--ease-out);
	}
	input:hover {
		border-color: var(--color-border-strong);
	}
	input:focus {
		outline: none;
		border-color: var(--color-primary);
		box-shadow: var(--shadow-focus);
	}
	input::placeholder {
		color: var(--color-text-disabled);
	}
</style>
