<script>
	// Empty state: icon + title + optional description + optional action slot.
	// `compact` tightens padding for use inside panels. Pass a default <svg> via
	// the `icon` snippet to override the generic glyph.
	let { title = '', description = '', compact = false, icon, children } = $props();
</script>

<div class="empty" class:compact>
	<span class="glyph" aria-hidden="true">
		{#if icon}
			{@render icon()}
		{:else}
			<svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
				<path d="M3 7a2 2 0 0 1 2-2h4l2 2h6a2 2 0 0 1 2 2v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z" />
			</svg>
		{/if}
	</span>
	<p class="title">{title}</p>
	{#if description}<p class="desc">{description}</p>{/if}
	{#if children}<div class="action">{@render children()}</div>{/if}
</div>

<style>
	.empty {
		display: flex;
		flex-direction: column;
		align-items: center;
		text-align: center;
		gap: var(--space-2);
		padding: var(--space-10) var(--space-6);
		color: var(--color-text-muted);
	}
	.empty.compact {
		padding: var(--space-6) var(--space-4);
	}
	.glyph {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 44px;
		height: 44px;
		margin-bottom: var(--space-1);
		border-radius: var(--radius-full);
		background-color: var(--color-surface-sunken);
		color: var(--color-text-muted);
	}
	.empty.compact .glyph {
		width: 38px;
		height: 38px;
	}
	.title {
		margin: 0;
		font-size: var(--text-base);
		font-weight: var(--weight-medium);
		color: var(--color-text-secondary);
	}
	.desc {
		margin: 0;
		max-width: 38ch;
		font-size: var(--text-sm);
		line-height: var(--leading-normal);
		color: var(--color-text-muted);
	}
	.action {
		margin-top: var(--space-3);
	}
</style>
