<script>
	// The Lumni lockup: 4-point spark mark + optional italic-serif wordmark + tag.
	// Color is controlled by the parent via `color` (wordmark) and CSS custom
	// props: --mark-fill (mark), --word-size, --tag-color. Defaults suit the dark
	// sidebar (Glow mark). On light surfaces, set --mark-fill to --color-primary.
	// `pulse` adds a gentle breathing glow to the spark mark.
	let { size = 22, wordmark = true, tag = '', pulse = false } = $props();
</script>

<span class="lockup">
	<svg class="mark" class:pulse viewBox="0 0 24 24" width={size} height={size} aria-hidden="true">
		<path d="M12 1.5 14.2 9.8 22.5 12 14.2 14.2 12 22.5 9.8 14.2 1.5 12 9.8 9.8Z" />
	</svg>
	{#if wordmark || tag}
		<span class="word-stack">
			{#if wordmark}<span class="word">mirante</span>{/if}
			{#if tag}<span class="tag">{tag}</span>{/if}
		</span>
	{/if}
</span>

<style>
	.lockup {
		display: inline-flex;
		align-items: center;
		gap: 7px;
	}
	.mark {
		fill: var(--mark-fill, var(--glow));
		flex-shrink: 0;
	}
	/* Optional gentle "breathing" pulse: a glow halo + subtle scale on the spark. */
	.mark.pulse {
		transform-origin: center;
		animation: mark-pulse 3s var(--ease-in-out) infinite;
	}
	@keyframes mark-pulse {
		0%,
		100% {
			transform: scale(1);
			filter: drop-shadow(0 0 1.5px color-mix(in srgb, var(--mark-fill, var(--glow)) 20%, transparent));
		}
		50% {
			transform: scale(1.06);
			filter: drop-shadow(0 0 7px color-mix(in srgb, var(--mark-fill, var(--glow)) 70%, transparent));
		}
	}
	/* Wordmark over tag (e.g. "mirante" with "by lumni" beneath it). */
	.word-stack {
		display: flex;
		flex-direction: column;
		gap: 2px;
	}
	.word {
		font-family: var(--font-serif);
		font-style: italic;
		font-weight: 400;
		font-size: var(--word-size, 16px);
		line-height: 1;
		color: currentColor;
		letter-spacing: -0.01em;
	}
	.tag {
		font-family: var(--font-mono);
		font-size: 9.5px;
		letter-spacing: var(--tracking-eyebrow);
		color: var(--tag-color, color-mix(in srgb, currentColor 45%, transparent));
	}
</style>
