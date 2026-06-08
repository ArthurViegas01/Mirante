// Runes-based session store, shared across components.

class SessionStore {
	user = $state(null);
	csrf = $state('');

	get authenticated() {
		return this.user !== null;
	}

	clear() {
		this.user = null;
		this.csrf = '';
	}
}

export const session = new SessionStore();
