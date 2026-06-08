<script>
	import { monitor } from '$lib/stores/monitor.svelte.js';

	let open = $state(false);

	async function toggle() {
		open = !open;
		if (open) await monitor.loadAlerts();
	}
</script>

<div class="nc">
	<button class="bell" onclick={toggle} aria-label="Notificações" title="Notificações">
		<svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round">
			<path d="M18 8a6 6 0 0 0-12 0c0 7-3 9-3 9h18s-3-2-3-9" />
			<path d="M13.73 21a2 2 0 0 1-3.46 0" />
		</svg>
		{#if monitor.unreadCount > 0}<span class="badge">{monitor.unreadCount}</span>{/if}
	</button>

	{#if open}
		<div class="panel">
			<header>
				<span class="h">Notificações</span>
				<button class="link" onclick={() => monitor.markAllRead()}>marcar todas lidas</button>
			</header>
			{#if monitor.alerts.length === 0}
				<p class="empty">Nada por aqui.</p>
			{:else}
				<ul>
					{#each monitor.alerts as a (a.id)}
						<li class:unread={!a.read_at}>
							<span class="dot {a.severity}" aria-hidden="true"></span>
							<div class="body">
								<span class="title">{a.title}</span>
								<span class="time">{a.created_at}</span>
							</div>
						</li>
					{/each}
				</ul>
			{/if}
		</div>
	{/if}
</div>

<style>
	.nc {
		position: relative;
	}
	.bell {
		position: relative;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 34px;
		height: 34px;
		border: var(--border-width-1) solid var(--color-border);
		border-radius: var(--radius-md);
		background-color: var(--color-surface);
		color: var(--color-text-muted);
		cursor: pointer;
	}
	.bell:hover {
		border-color: var(--color-border-strong);
		color: var(--color-text);
	}
	.bell:focus-visible {
		outline: none;
		box-shadow: var(--shadow-focus);
	}
	.badge {
		position: absolute;
		top: -5px;
		right: -5px;
		min-width: 16px;
		height: 16px;
		padding: 0 4px;
		border-radius: var(--radius-full);
		background-color: var(--color-danger);
		color: #fff;
		font-family: var(--font-mono);
		font-size: 10px;
		line-height: 16px;
		text-align: center;
	}
	.panel {
		position: absolute;
		top: 42px;
		right: 0;
		width: 320px;
		max-height: 420px;
		overflow-y: auto;
		background-color: var(--color-surface-elevated);
		border: var(--border-width-1) solid var(--color-border);
		border-radius: var(--radius-lg);
		box-shadow: var(--shadow-lg);
		z-index: 20;
	}
	header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 12px 16px;
		border-bottom: var(--border-width-1) solid var(--color-divider);
	}
	.h {
		font-weight: var(--weight-semibold);
		font-size: var(--text-sm);
	}
	.link {
		background: none;
		border: none;
		color: var(--color-link);
		font-size: var(--text-xs);
		cursor: pointer;
	}
	.empty {
		padding: var(--space-6);
		text-align: center;
		color: var(--color-text-muted);
		font-size: var(--text-sm);
		margin: 0;
	}
	ul {
		list-style: none;
		margin: 0;
		padding: 0;
	}
	li {
		display: flex;
		gap: 10px;
		padding: 12px 16px;
		border-bottom: var(--border-width-1) solid var(--color-divider);
	}
	li:last-child {
		border-bottom: none;
	}
	li.unread {
		background-color: color-mix(in srgb, var(--color-accent) 6%, transparent);
	}
	.dot {
		width: 8px;
		height: 8px;
		border-radius: var(--radius-full);
		margin-top: 5px;
		flex-shrink: 0;
	}
	.dot.success {
		background-color: var(--color-success);
	}
	.dot.warning {
		background-color: var(--color-warning);
	}
	.dot.danger {
		background-color: var(--color-danger);
	}
	.dot.info {
		background-color: var(--color-info);
	}
	.body {
		display: flex;
		flex-direction: column;
		gap: 2px;
	}
	.title {
		font-size: 13px;
		color: var(--color-text);
	}
	.time {
		font-family: var(--font-mono);
		font-size: 11px;
		color: var(--color-text-muted);
	}
</style>
