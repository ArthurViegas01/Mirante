<script>
	// points: latency numbers, oldest → newest. The last point gets the Glow
	// "live" dot + halo (the Mirante real-time signature).
	let { points = [] } = $props();

	const W = 600;
	const H = 120;
	const PAD = 8;

	let path = $derived(buildPath(points));
	let last = $derived(lastPoint(points));

	function coords(pts) {
		const max = Math.max(...pts, 1);
		const n = pts.length;
		return pts.map((v, i) => ({
			x: n <= 1 ? W : (i / (n - 1)) * W,
			y: H - (v / max) * (H - 2 * PAD) - PAD
		}));
	}
	function buildPath(pts) {
		if (!pts.length) return '';
		return coords(pts)
			.map((p, i) => `${i === 0 ? 'M' : 'L'} ${p.x.toFixed(1)} ${p.y.toFixed(1)}`)
			.join(' ');
	}
	function lastPoint(pts) {
		if (!pts.length) return null;
		const c = coords(pts);
		return c[c.length - 1];
	}
</script>

{#if path}
	<svg class="spark" viewBox="0 0 {W} {H}" preserveAspectRatio="none" role="img" aria-label="Latência recente">
		<path
			d={path}
			fill="none"
			stroke="var(--chart-1)"
			stroke-width="2.2"
			vector-effect="non-scaling-stroke"
			stroke-linejoin="round"
			stroke-linecap="round"
		/>
		{#if last}
			<circle cx={last.x} cy={last.y} r="9" fill="var(--color-live-halo)" />
			<circle cx={last.x} cy={last.y} r="4.5" fill="var(--color-live)" />
		{/if}
	</svg>
{:else}
	<p class="empty">Sem dados ainda.</p>
{/if}

<style>
	.spark {
		width: 100%;
		height: 120px;
		display: block;
	}
	.empty {
		color: var(--color-text-muted);
		font-size: var(--text-sm);
		margin: 0;
	}
</style>
