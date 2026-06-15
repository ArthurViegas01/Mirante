<script>
	let {
		value = $bindable(''),
		label = '',
		placeholder = '',
		id = '',
		name = '',
		rows = 3,
		required = false,
		disabled = false,
		error = '',
		hint = ''
	} = $props();
</script>

<label class="field" class:has-error={!!error}>
	{#if label}<span class="label">{label}</span>{/if}
	<textarea
		{id}
		{name}
		{placeholder}
		{rows}
		{required}
		{disabled}
		{value}
		aria-invalid={error ? 'true' : undefined}
		oninput={(e) => (value = e.currentTarget.value)}
	></textarea>
	{#if error}
		<span class="msg err" role="alert">{error}</span>
	{:else if hint}
		<span class="msg hint">{hint}</span>
	{/if}
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
	textarea {
		width: 100%;
		font-family: var(--font-sans);
		font-size: 13px;
		line-height: var(--leading-normal);
		padding: 8px 10px;
		background-color: var(--color-surface);
		color: var(--color-text);
		border: var(--border-width-1) solid var(--color-border);
		border-radius: var(--radius-md);
		resize: vertical;
		min-height: 64px;
		transition:
			border-color var(--dur-fast) var(--ease-out),
			box-shadow var(--dur-fast) var(--ease-out);
	}
	textarea:hover:not(:disabled) {
		border-color: var(--color-border-strong);
	}
	textarea:focus {
		outline: none;
		border-color: var(--color-primary);
		box-shadow: var(--shadow-focus);
	}
	textarea::placeholder {
		color: var(--color-text-disabled);
	}
	textarea:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}
	.has-error textarea {
		border-color: var(--color-danger);
	}
	.has-error textarea:focus {
		box-shadow: 0 0 0 3px color-mix(in srgb, var(--color-danger) 22%, transparent);
	}
	.msg {
		font-size: var(--text-xs);
	}
	.msg.err {
		color: var(--color-danger-text);
	}
	.msg.hint {
		color: var(--color-text-muted);
	}
</style>
