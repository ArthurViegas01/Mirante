// Promise-based confirmation dialog. Mount <ConfirmHost /> once (root layout);
// call `await confirm.ask({ ... })` anywhere to replace the native confirm().

class ConfirmStore {
	open = $state(false);
	opts = $state({});
	#resolve = null;

	ask(opts = {}) {
		this.opts = {
			title: 'Confirmar',
			message: '',
			confirmLabel: 'Confirmar',
			cancelLabel: 'Cancelar',
			danger: false,
			...opts
		};
		this.open = true;
		return new Promise((resolve) => {
			this.#resolve = resolve;
		});
	}

	#settle(value) {
		this.open = false;
		if (this.#resolve) {
			this.#resolve(value);
			this.#resolve = null;
		}
	}

	confirm() {
		this.#settle(true);
	}

	cancel() {
		this.#settle(false);
	}
}

export const confirm = new ConfirmStore();
