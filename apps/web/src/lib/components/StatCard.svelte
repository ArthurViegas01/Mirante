<script>
	// KPI card (Lumni DS 7.2). Becomes a link when `href` is set. `live` adds the
	// Glow real-time dot; `tone` colors the hint line (success/warning/danger).
	let { label = '', value = '', hint = '', href = null, live = false, tone = 'default' } =
		$props();
</script>

<svelte:element
	this={href ? 'a' : 'div'}
	{href}
	class="stat tone-{tone}"
	class:link={!!href}
	class:live
>
	<span class="label">{label}</span>
	{#if live}<span class="live-dot" aria-hidden="true"></span>{/if}
	<span class="value tnum">{value}</span>
	{#if hint}<span class="hint">{hint}</span>{/if}
</svelte:element>

<style>
	.stat {
		position: relative;
		display: flex;
		flex-direction: column;
		padding: 18px 20px;
		background-color: var(--color-surface);
		border: var(--border-width-1) solid var(--color-border);
		border-radius: var(--radius-lg);
		box-shadow: var(--shadow-sm);
		text-decoration: none;
		color: inherit;
	}
	.stat.link {
		transition:
			border-color var(--dur-fast) var(--ease-out),
			box-shadow var(--dur-fast) var(--ease-out),
			transform var(--dur-fast) var(--ease-out);
	}
	.stat.link:hover {
		border-color: var(--color-border-strong);
		box-shadow: var(--shadow-md);
		transform: translateY(-2px);
	}
	.stat.link:focus-visible {
		outline: none;
		box-shadow: var(--shadow-focus);
	}
	.label {
		font-family: var(--font-mono);
		font-size: 11px;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		color: var(--color-text-muted);
	}
	.value {
		margin-top: 8px;
		font-family: var(--font-mono);
		font-size: 28px;
		font-weight: var(--weight-medium);
		line-height: 1;
		letter-spacing: -0.015em;
		color: var(--color-text);
	}
	.hint {
		margin-top: 8px;
		font-family: var(--font-mono);
		font-size: 11.5px;
		color: var(--color-text-muted);
	}
	.tone-success .hint {
		color: var(--color-success-text);
	}
	.tone-warning .hint {
		color: var(--color-warning-text);
	}
	.tone-danger .hint {
		color: var(--color-danger-text);
	}
	.live-dot {
		position: absolute;
		top: 18px;
		right: 18px;
		width: 8px;
		height: 8px;
		border-radius: var(--radius-full);
		background-color: var(--color-live);
		box-shadow: 0 0 0 4px var(--color-live-halo);
		animation: live-pulse 2.4s var(--ease-in-out) infinite;
	}
	@keyframes live-pulse {
		0%,
		100% {
			box-shadow: 0 0 0 3px var(--color-live-halo);
		}
		50% {
			box-shadow: 0 0 0 7px color-mix(in srgb, var(--glow) 8%, transparent);
		}
	}
	@media (prefers-reduced-motion: reduce) {
		.live-dot {
			animation: none;
		}
	}
</style>
