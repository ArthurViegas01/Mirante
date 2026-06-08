<script>
	import { onMount } from 'svelte';

	let theme = $state('light');

	onMount(() => {
		theme = document.documentElement.dataset.theme || 'light';
	});

	function toggle() {
		theme = theme === 'dark' ? 'light' : 'dark';
		document.documentElement.dataset.theme = theme;
		try {
			localStorage.setItem('theme', theme);
		} catch (e) {
			/* ignore */
		}
	}
</script>

<button class="toggle" onclick={toggle} aria-label="Alternar tema" title="Alternar tema">
	{#if theme === 'dark'}
		<!-- moon -->
		<svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round">
			<path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z" />
		</svg>
	{:else}
		<!-- sun -->
		<svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round">
			<circle cx="12" cy="12" r="4" />
			<path d="M12 2v2M12 20v2M4.93 4.93l1.41 1.41M17.66 17.66l1.41 1.41M2 12h2M20 12h2M6.34 17.66l-1.41 1.41M19.07 4.93l-1.41 1.41" />
		</svg>
	{/if}
</button>

<style>
	.toggle {
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
		transition:
			border-color var(--dur-fast) var(--ease-out),
			color var(--dur-fast) var(--ease-out);
	}
	.toggle:hover {
		border-color: var(--color-border-strong);
		color: var(--color-text);
	}
	.toggle:focus-visible {
		outline: none;
		box-shadow: var(--shadow-focus);
	}
</style>
