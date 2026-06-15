<script>
	import { onMount } from 'svelte';
	import StatCard from '$lib/components/StatCard.svelte';
	import StatusBadge from '$lib/components/StatusBadge.svelte';
	import Skeleton from '$lib/components/Skeleton.svelte';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import { api } from '$lib/api.js';
	import { session } from '$lib/stores/session.svelte.js';
	import { monitor } from '$lib/stores/monitor.svelte.js';
	import { svcVariant, svcLabel } from '$lib/serviceStatus.js';
	import { isOverdue, prazoLabel, daysUntil } from '$lib/taskMeta.js';
	import { sumByCurrency, formatMoney } from '$lib/money.js';
	import { APP_PIPELINE, APP_STATUS } from '$lib/applicationMeta.js';

	let loading = $state(true);
	let error = $state('');
	let projects = $state([]);
	let tasks = $state([]);
	let subs = $state([]);
	let services = $state([]);
	let apps = $state([]);

	onMount(async () => {
		try {
			const [p, t, s, sv, a] = await Promise.all([
				api('/api/projects'),
				api('/api/tasks'),
				api('/api/subscriptions'),
				api('/api/services'),
				api('/api/applications')
			]);
			projects = p.projects ?? [];
			tasks = t.tasks ?? [];
			subs = s.subscriptions ?? [];
			services = sv.services ?? [];
			apps = a.applications ?? [];
		} catch (e) {
			error = e.message;
		} finally {
			loading = false;
		}
	});

	const statusRank = { down: 0, degraded: 1, unknown: 2, paused: 3, up: 4 };
	const projName = (id) => projects.find((p) => p.id === id)?.nome ?? '';

	let activeProjects = $derived(
		projects.filter((p) => p.status === 'ativo' || p.status === 'no_ar')
	);
	let openTasks = $derived(tasks.filter((t) => t.status !== 'feito'));
	let overdueTasks = $derived(tasks.filter((t) => isOverdue(t.prazo, t.status)));
	let focusTasks = $derived(
		[...openTasks]
			.sort((a, b) => {
				const da = daysUntil(a.prazo);
				const db = daysUntil(b.prazo);
				if (da === null && db === null) return 0;
				if (da === null) return 1;
				if (db === null) return -1;
				return da - db;
			})
			.slice(0, 6)
	);
	let servicesUp = $derived(services.filter((s) => s.status === 'up').length);
	let servicesBad = $derived(
		services.filter((s) => s.status === 'down' || s.status === 'degraded').length
	);
	let allServicesUp = $derived(services.length > 0 && servicesUp === services.length);
	let sortedServices = $derived(
		[...services]
			.sort((a, b) => (statusRank[a.status] ?? 9) - (statusRank[b.status] ?? 9))
			.slice(0, 7)
	);
	let totals = $derived(sumByCurrency(subs));
	let costEntries = $derived(Object.entries(totals).filter(([, c]) => c > 0));
	let activeApps = $derived(apps.filter((a) => a.status !== 'rejeitado' && a.status !== 'aceito'));
	let pipeline = $derived(
		APP_PIPELINE.map((st) => ({ st, n: apps.filter((a) => a.status === st).length }))
	);
	let recentAlerts = $derived(monitor.alerts.slice(0, 6));

	function greeting() {
		const h = new Date().getHours();
		return h < 12 ? 'Bom dia' : h < 18 ? 'Boa tarde' : 'Boa noite';
	}
	const dateLabel = new Date().toLocaleDateString('pt-BR', {
		weekday: 'long',
		day: 'numeric',
		month: 'long'
	});
</script>

<header class="head">
	<p class="eyebrow">Mirante</p>
	<h1>{greeting()}, {session.displayName}<span class="dot">.</span></h1>
	<p class="sub">{dateLabel}</p>
</header>

{#if error}
	<div class="panel"><p class="err">{error}</p></div>
{:else if loading}
	<section class="kpis">
		{#each Array(5) as _}
			<div class="sk-card">
				<Skeleton w="60%" h="11px" />
				<Skeleton w="45%" h="26px" />
				<Skeleton w="50%" h="11px" />
			</div>
		{/each}
	</section>
	<div class="grid">
		{#each Array(2) as _}
			<div class="panel">
				<div class="panel-head"><Skeleton w="120px" h="14px" /></div>
				<div class="panel-body sk-rows">
					{#each Array(4) as __}<Skeleton w="100%" h="34px" radius="var(--radius-md)" />{/each}
				</div>
			</div>
		{/each}
	</div>
{:else}
	<!-- KPIs -->
	<section class="kpis">
		<StatCard
			label="Projetos ativos"
			value={activeProjects.length}
			hint={`${projects.length} no total`}
			href="/projetos"
		/>
		<StatCard
			label="Serviços no ar"
			value={`${servicesUp}/${services.length}`}
			live={allServicesUp}
			tone={servicesBad > 0 ? 'danger' : allServicesUp ? 'success' : 'default'}
			hint={services.length === 0
				? 'nenhum monitorado'
				: servicesBad > 0
					? `${servicesBad} com problema`
					: allServicesUp
						? 'tudo no ar'
						: `${servicesUp} no ar`}
			href="/projetos"
		/>
		<StatCard
			label="Tarefas abertas"
			value={openTasks.length}
			tone={overdueTasks.length > 0 ? 'danger' : 'default'}
			hint={overdueTasks.length > 0
				? `${overdueTasks.length} atrasada${overdueTasks.length > 1 ? 's' : ''}`
				: 'em dia'}
			href="/tarefas"
		/>
		<StatCard
			label="Custo mensal"
			value={costEntries.length ? formatMoney(costEntries[0][1], costEntries[0][0]) : 'R$ 0,00'}
			hint={costEntries.length > 1
				? '+ ' + costEntries.slice(1).map(([m, c]) => formatMoney(c, m)).join(' + ')
				: 'por mês'}
			href="/custos"
		/>
		<StatCard
			label="No pipeline"
			value={activeApps.length}
			hint={`${apps.length} candidatura${apps.length === 1 ? '' : 's'}`}
			href="/candidaturas"
		/>
	</section>

	<div class="grid">
		<!-- Ao vivo: serviços -->
		<section class="panel">
			<div class="panel-head">
				<span class="panel-title">Ao vivo</span>
				<a class="panel-link" href="/projetos">Projetos →</a>
			</div>
			{#if sortedServices.length === 0}
				<EmptyState
					compact
					title="Nenhum serviço monitorado"
					description="Adicione serviços a um projeto para acompanhar a saúde em tempo real."
				/>
			{:else}
				<ul class="rows">
					{#each sortedServices as s (s.id)}
						<li class="row">
							<span class="row-main">
								<span class="row-title">{s.nome}</span>
								{#if projName(s.project_id)}
									<a class="row-meta" href={`/projetos/${s.project_id}`}>{projName(s.project_id)}</a>
								{/if}
							</span>
							<StatusBadge status={svcVariant(s.status)} label={svcLabel(s.status)} />
						</li>
					{/each}
				</ul>
			{/if}
		</section>

		<!-- Foco: tarefas -->
		<section class="panel">
			<div class="panel-head">
				<span class="panel-title">Foco</span>
				<a class="panel-link" href="/tarefas">Quadro →</a>
			</div>
			{#if focusTasks.length === 0}
				<EmptyState compact title="Sem tarefas abertas" description="Tudo em dia por aqui." />
			{:else}
				<ul class="rows">
					{#each focusTasks as t (t.id)}
						<li class="row">
							<span class="row-main">
								<span class="row-title">{t.titulo}</span>
								{#if projName(t.project_id)}
									<span class="row-meta">{projName(t.project_id)}</span>
								{/if}
							</span>
							{#if t.prazo}
								<span class="prazo" class:late={isOverdue(t.prazo, t.status)}>
									{prazoLabel(t.prazo)}
								</span>
							{/if}
						</li>
					{/each}
				</ul>
			{/if}
		</section>

		<!-- Atividade recente: alertas -->
		<section class="panel">
			<div class="panel-head">
				<span class="panel-title">Atividade recente</span>
			</div>
			{#if recentAlerts.length === 0}
				<EmptyState compact title="Nada por aqui ainda" description="Alertas do monitor aparecem aqui." />
			{:else}
				<ul class="rows">
					{#each recentAlerts as a (a.id)}
						<li class="row alert">
							<span class="sev {a.severity}" aria-hidden="true"></span>
							<span class="row-main">
								<span class="row-title">{a.title}</span>
								<span class="row-meta mono">{a.created_at}</span>
							</span>
						</li>
					{/each}
				</ul>
			{/if}
		</section>

		<!-- Carreira: pipeline -->
		<section class="panel">
			<div class="panel-head">
				<span class="panel-title">Carreira</span>
				<a class="panel-link" href="/vagas">Vagas →</a>
			</div>
			{#if apps.length === 0}
				<EmptyState compact title="Pipeline vazio" description="Acompanhe vagas em Candidaturas." />
			{:else}
				<div class="pipeline">
					{#each pipeline as { st, n } (st)}
						<a class="stage" class:zero={n === 0} href="/candidaturas">
							<span class="stage-n tnum">{n}</span>
							<span class="stage-l">{APP_STATUS[st].label}</span>
						</a>
					{/each}
				</div>
			{/if}
		</section>
	</div>
{/if}

<style>
	.head {
		margin-bottom: var(--space-8);
	}
	.eyebrow {
		margin: 0 0 var(--space-2);
	}
	h1 {
		font-size: var(--text-3xl);
		font-weight: var(--weight-medium);
		letter-spacing: var(--tracking-tight);
		line-height: var(--leading-tight);
		color: var(--color-text);
		margin: 0;
		text-transform: capitalize;
	}
	h1 .dot {
		color: var(--color-accent);
	}
	.sub {
		margin: var(--space-2) 0 0;
		font-size: var(--text-sm);
		color: var(--color-text-muted);
		text-transform: capitalize;
	}

	.kpis {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
		gap: var(--space-4);
		margin-bottom: var(--space-6);
		animation: rise var(--dur-slow) var(--ease-out) both;
	}
	@keyframes rise {
		from {
			opacity: 0;
			transform: translateY(8px);
		}
		to {
			opacity: 1;
			transform: none;
		}
	}
	.sk-card {
		display: flex;
		flex-direction: column;
		gap: 10px;
		padding: 18px 20px;
		background-color: var(--color-surface);
		border: var(--border-width-1) solid var(--color-border);
		border-radius: var(--radius-lg);
	}

	.grid {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: var(--space-4);
		margin-bottom: var(--space-4);
		animation: rise var(--dur-slow) var(--ease-out) both;
		animation-delay: 60ms;
	}
	@media (max-width: 880px) {
		.grid {
			grid-template-columns: 1fr;
		}
	}

	.panel {
		background-color: var(--color-surface);
		border: var(--border-width-1) solid var(--color-border);
		border-radius: var(--radius-lg);
		box-shadow: var(--shadow-sm);
		overflow: hidden;
	}
	.panel-head {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 14px 18px;
		border-bottom: var(--border-width-1) solid var(--color-divider);
	}
	.panel-title {
		font-size: 14px;
		font-weight: var(--weight-semibold);
		color: var(--color-text);
	}
	.panel-link {
		font-size: var(--text-xs);
		font-family: var(--font-mono);
		color: var(--color-link);
	}
	.panel-body {
		padding: var(--space-4) var(--space-5);
	}
	.sk-rows {
		display: flex;
		flex-direction: column;
		gap: 10px;
	}
	.err {
		padding: var(--space-6);
		margin: 0;
		color: var(--color-danger-text);
	}

	.rows {
		list-style: none;
		margin: 0;
		padding: 0;
	}
	.row {
		display: flex;
		align-items: center;
		gap: var(--space-3);
		justify-content: space-between;
		padding: 11px 18px;
		border-bottom: var(--border-width-1) solid var(--color-divider);
	}
	.row:last-child {
		border-bottom: none;
	}
	.row-main {
		display: flex;
		flex-direction: column;
		gap: 1px;
		min-width: 0;
	}
	.row-title {
		font-size: 13.5px;
		color: var(--color-text);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.row-meta {
		font-size: 12px;
		color: var(--color-text-muted);
	}
	.row-meta.mono {
		font-family: var(--font-mono);
		font-size: 11px;
	}
	a.row-meta:hover {
		color: var(--color-link);
	}
	.prazo {
		flex-shrink: 0;
		font-family: var(--font-mono);
		font-size: 11.5px;
		color: var(--color-text-muted);
	}
	.prazo.late {
		color: var(--color-danger-text);
	}

	.row.alert {
		justify-content: flex-start;
	}
	.sev {
		width: 8px;
		height: 8px;
		flex-shrink: 0;
		margin-top: 5px;
		align-self: flex-start;
		border-radius: var(--radius-full);
	}
	.sev.success {
		background-color: var(--color-success);
	}
	.sev.warning {
		background-color: var(--color-warning);
	}
	.sev.danger {
		background-color: var(--color-danger);
	}
	.sev.info {
		background-color: var(--color-info);
	}

	.pipeline {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: var(--space-2);
		padding: var(--space-4) var(--space-5);
	}
	.stage {
		display: flex;
		flex-direction: column;
		gap: 2px;
		padding: 10px 12px;
		border-radius: var(--radius-md);
		background-color: var(--color-surface-sunken);
		text-decoration: none;
		transition: background-color var(--dur-fast) var(--ease-out);
	}
	.stage:hover {
		background-color: color-mix(in srgb, var(--color-accent) 8%, var(--color-surface-sunken));
	}
	.stage.zero {
		opacity: 0.5;
	}
	.stage-n {
		font-family: var(--font-mono);
		font-size: 20px;
		font-weight: var(--weight-medium);
		color: var(--color-text);
		line-height: 1;
	}
	.stage-l {
		font-size: 11.5px;
		color: var(--color-text-muted);
	}
</style>
