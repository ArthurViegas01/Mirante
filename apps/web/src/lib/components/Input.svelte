<script>
	// `type` is intentionally applied one-way (not bound) so the input can carry a
	// dynamic type while still supporting two-way `value` via $bindable.
	let {
		value = $bindable(''),
		type = 'text',
		label = '',
		placeholder = '',
		id = '',
		name = '',
		required = false,
		disabled = false,
		autocomplete = undefined,
		inputmode = undefined,
		error = '',
		hint = ''
	} = $props();

	let revealed = $state(false);
	let isPassword = $derived(type === 'password');
	let effectiveType = $derived(isPassword && revealed ? 'text' : type);
</script>

<label class="field" class:has-error={!!error}>
	{#if label}<span class="label">{label}</span>{/if}
	<span class="control">
		<input
			{id}
			{name}
			{placeholder}
			{required}
			{disabled}
			{autocomplete}
			{inputmode}
			{value}
			type={effectiveType}
			class:with-affix={isPassword}
			aria-invalid={error ? 'true' : undefined}
			oninput={(e) => (value = e.currentTarget.value)}
		/>
		{#if isPassword}
			<button
				type="button"
				class="reveal"
				tabindex="-1"
				onclick={() => (revealed = !revealed)}
				aria-label={revealed ? 'Ocultar senha' : 'Mostrar senha'}
				title={revealed ? 'Ocultar senha' : 'Mostrar senha'}
			>
				{#if revealed}
					<svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round">
						<path d="M9.88 9.88a3 3 0 0 0 4.24 4.24" />
						<path d="M10.73 5.08A10.4 10.4 0 0 1 12 5c6 0 10 7 10 7a18.5 18.5 0 0 1-2.16 3.19" />
						<path d="M6.61 6.61A18 18 0 0 0 2 12s4 7 10 7a9.7 9.7 0 0 0 5.39-1.61" />
						<path d="m2 2 20 20" />
					</svg>
				{:else}
					<svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round">
						<path d="M2 12s4-7 10-7 10 7 10 7-4 7-10 7-10-7-10-7Z" />
						<circle cx="12" cy="12" r="3" />
					</svg>
				{/if}
			</button>
		{/if}
	</span>
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
	.control {
		position: relative;
		display: flex;
	}
	input {
		width: 100%;
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
	input.with-affix {
		padding-right: 38px;
	}
	input:hover:not(:disabled) {
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
	input:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}
	.has-error input {
		border-color: var(--color-danger);
	}
	.has-error input:focus {
		box-shadow: 0 0 0 3px color-mix(in srgb, var(--color-danger) 22%, transparent);
	}
	.reveal {
		position: absolute;
		top: 50%;
		right: 6px;
		transform: translateY(-50%);
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 26px;
		height: 26px;
		border: none;
		border-radius: var(--radius-sm);
		background: transparent;
		color: var(--color-text-muted);
		cursor: pointer;
		transition: color var(--dur-fast) var(--ease-out);
	}
	.reveal:hover {
		color: var(--color-text);
	}
	.reveal:focus-visible {
		outline: none;
		box-shadow: var(--shadow-focus);
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
