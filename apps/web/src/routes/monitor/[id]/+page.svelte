<script>
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import Button from '$lib/components/Button.svelte';
	import StatusBadge from '$lib/components/StatusBadge.svelte';
	import ProgressBar from '$lib/components/ProgressBar.svelte';
	import Sparkline from '$lib/components/Sparkline.svelte';
	import { api } from '$lib/api.js';
	import { monitor } from '$lib/stores/monitor.svelte.js';
	import { svcVariant, svcLabel, uptimeVariant, camadaLabel } from '$lib/serviceStatus.js';

	let id = $derived($page.params.id);
	let detail = $state(null);
	let loading = $state(true);
	let error = $state('');

	async function load(sid) {
		loading = true;
		error = '';
		try {
			detail = await api(`/api/services/${sid}`);
		} catch (e) {
			error = e.message;
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		load(id);
	});

	// Live: reload when a transition for this service arrives.
	$effect(() => {
		const ev = monitor.lastEvent;
		if (ev && ev.service_id === id) load(id);
	});

	// Checks come newest-first; reverse to oldest→newest latency for the sparkline.
	let points = $derived(detail ? detail.checks.slice().reverse().map((c) => c.latency_ms || 0) : []);

	function pct(u) {
		return u.samples > 0 ? `${(u.up_ratio * 100).toFixed(1)}%` : 'sem dados';
	}

	async function toggleEnabled() {
		await api(`/api/services/${id}/enabled`, {
			method: 'POST',
			body: { enabled: !detail.service.enabled }
		});
		await load(id);
	}

	async function remove() {
		if (!confirm('Excluir este serviço e seu histórico?')) return;
		await api(`/api/services/${id}`, { method: 'DELETE' });
		goto('/monitor');
	}
</script>

<a class="back" href="/monitor">← Monitor</a>

{#if loading && !detail}
	<p class="muted">Carregando…</p>
{:else if error}
	<p class="error">{error}</p>
{:else if detail}
	{@const s = detail.service}
	<header class="head">
		<div>
			<h1>{s.nome}</h1>
			<span class="target">{s.kind} · {s.target}</span>
		</div>
		<StatusBadge status={svcVariant(s.current_status)} label={svcLabel(s.current_status)} />
	</header>

	<section class="panel">
		<h2>Latência recente</h2>
		<Sparkline {points} />
	</section>

	<section class="panel">
		<h2>Uptime</h2>
		<div class="uptime">
			{#each [['24h', detail.uptime_24h], ['7 dias', detail.uptime_7d], ['30 dias', detail.uptime_30d]] as [label, u] (label)}
				<div class="up">
					<div class="up-head"><span>{label}</span><span class="mono">{pct(u)}</span></div>
					<ProgressBar value={u.up_ratio} variant={uptimeVariant(u.up_ratio)} />
				</div>
			{/each}
		</div>
	</section>

	<section class="panel">
		<h2>Configuração</h2>
		<dl class="config">
			{#if s.camada}<div><dt>Camada</dt><dd>{camadaLabel(s.camada)}</dd></div>{/if}
			{#if s.provider}<div><dt>Provedor</dt><dd>{s.provider}</dd></div>{/if}
			<div><dt>Intervalo</dt><dd>{s.interval_seconds}s</dd></div>
			<div><dt>Timeout</dt><dd>{s.timeout_ms}ms</dd></div>
			<div><dt>Degradado &gt;</dt><dd>{s.degraded_threshold_ms}ms</dd></div>
			<div><dt>Anti-flap</dt><dd>{s.anti_flap_n} falhas / {s.recovery_k} sucessos</dd></div>
			{#if s.kind === 'http'}<div><dt>Status esperado</dt><dd>{s.expected_status}</dd></div>{/if}
		</dl>
	</section>

	<section class="actions">
		<Button variant="secondary" onclick={toggleEnabled}>
			{s.enabled ? 'Pausar' : 'Retomar'}
		</Button>
		<Button variant="ghost" onclick={remove}>Excluir serviço</Button>
	</section>
{/if}

<style>
	.back {
		font-size: var(--text-sm);
		color: var(--color-text-muted);
		text-decoration: none;
		display: inline-block;
		margin-bottom: var(--space-4);
	}
	.back:hover {
		color: var(--color-text);
	}
	.head {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: var(--space-4);
		margin-bottom: var(--space-6);
	}
	h1 {
		font-size: var(--text-2xl);
		font-weight: var(--weight-medium);
		letter-spacing: var(--tracking-snug);
		color: var(--color-text);
		margin: 0 0 4px;
	}
	.target {
		font-family: var(--font-mono);
		font-size: var(--text-sm);
		color: var(--color-text-muted);
	}
	.panel {
		background-color: var(--color-surface);
		border: var(--border-width-1) solid var(--color-border);
		border-radius: var(--radius-lg);
		box-shadow: var(--shadow-sm);
		padding: var(--space-6);
		margin-bottom: var(--space-5);
	}
	h2 {
		font-size: var(--text-lg);
		font-weight: var(--weight-medium);
		color: var(--color-text);
		margin: 0 0 var(--space-4);
	}
	.uptime {
		display: flex;
		flex-direction: column;
		gap: var(--space-4);
	}
	.up-head {
		display: flex;
		justify-content: space-between;
		font-size: var(--text-sm);
		color: var(--color-text-secondary);
		margin-bottom: 6px;
	}
	.mono {
		font-family: var(--font-mono);
		font-feature-settings: 'tnum' 1;
		color: var(--color-text);
	}
	.config {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
		gap: var(--space-4);
		margin: 0;
	}
	.config dt {
		font-size: var(--text-xs);
		text-transform: uppercase;
		letter-spacing: 0.06em;
		color: var(--color-text-muted);
		margin-bottom: 2px;
	}
	.config dd {
		margin: 0;
		font-family: var(--font-mono);
		font-size: 13px;
		color: var(--color-text);
	}
	.actions {
		display: flex;
		gap: var(--space-3);
		margin-top: var(--space-6);
	}
	.muted {
		color: var(--color-text-secondary);
	}
	.error {
		color: var(--color-danger-text);
		font-size: var(--text-sm);
	}
</style>
