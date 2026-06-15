// App-wide toast queue. Mount <Toaster /> once (in the root layout); call
// toasts.success/error/info from anywhere to surface action feedback.

class ToastStore {
	items = $state([]);
	#seq = 0;

	push(message, { type = 'info', duration = 4000 } = {}) {
		const id = ++this.#seq;
		this.items = [...this.items, { id, message, type }];
		if (duration > 0) setTimeout(() => this.dismiss(id), duration);
		return id;
	}

	success(message, opts) {
		return this.push(message, { ...opts, type: 'success' });
	}

	error(message, opts) {
		// Errors linger a little longer by default.
		return this.push(message, { duration: 6000, ...opts, type: 'error' });
	}

	info(message, opts) {
		return this.push(message, { ...opts, type: 'info' });
	}

	dismiss(id) {
		this.items = this.items.filter((t) => t.id !== id);
	}
}

export const toasts = new ToastStore();
