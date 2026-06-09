// Runes-based session store, shared across components.

class SessionStore {
	user = $state(null);
	csrf = $state('');
	// True when the instance has no owner yet → route anonymous visitors to
	// /signup instead of /login. Resolved from GET /api/auth/status on boot.
	needsSetup = $state(false);

	get authenticated() {
		return this.user !== null;
	}

	clear() {
		this.user = null;
		this.csrf = '';
		// If you were logged in, an owner exists, so setup is not needed.
		this.needsSetup = false;
	}
}

export const session = new SessionStore();
