<script>
	import Modal from '$lib/components/Modal.svelte';
	import Button from '$lib/components/Button.svelte';
	import { confirm } from '$lib/stores/confirm.svelte.js';
</script>

<Modal
	bind:open={confirm.open}
	size="sm"
	title={confirm.opts.title}
	onclose={() => confirm.cancel()}
>
	{#if confirm.opts.message}<p class="msg">{confirm.opts.message}</p>{/if}
	{#snippet footer()}
		<Button variant="secondary" onclick={() => confirm.cancel()}>
			{confirm.opts.cancelLabel}
		</Button>
		<Button variant={confirm.opts.danger ? 'danger' : 'primary'} onclick={() => confirm.confirm()}>
			{confirm.opts.confirmLabel}
		</Button>
	{/snippet}
</Modal>

<style>
	.msg {
		margin: 0;
		font-size: var(--text-base);
		line-height: var(--leading-relaxed);
		color: var(--color-text-secondary);
	}
</style>
