<script>
	// Shimmer placeholder. Size via props; collapses to a static block under
	// prefers-reduced-motion.
	let { w = '100%', h = '1em', radius = 'var(--radius-sm)', block = false } = $props();
</script>

<span class="sk" class:block style="--w:{w}; --h:{h}; --r:{radius}"></span>

<style>
	.sk {
		display: inline-block;
		width: var(--w);
		height: var(--h);
		border-radius: var(--r);
		background-color: var(--color-surface-sunken);
		background-image: linear-gradient(
			100deg,
			transparent 30%,
			color-mix(in srgb, var(--color-border) 55%, transparent) 50%,
			transparent 70%
		);
		background-size: 200% 100%;
		background-repeat: no-repeat;
		animation: sk-shimmer 1.3s var(--ease-in-out) infinite;
	}
	.sk.block {
		display: block;
	}
	@keyframes sk-shimmer {
		from {
			background-position: 200% 0;
		}
		to {
			background-position: -200% 0;
		}
	}
	@media (prefers-reduced-motion: reduce) {
		.sk {
			animation: none;
		}
	}
</style>
