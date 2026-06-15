// Runes-based session store, shared across components.

class SessionStore {
	user = $state(null);
	csrf = $state('');
	// True when the instance has no owner yet (resolved from GET /api/auth/status
	// on boot). The signup screen uses it to show the form vs. a "closed" notice.
	needsSetup = $state(false);

	get authenticated() {
		return this.user !== null;
	}

	// Friendly label for the owner: name, else the local-part of the e-mail.
	get displayName() {
		if (!this.user) return '';
		return this.user.name?.trim() || this.user.email?.split('@')[0] || 'Você';
	}

	// One or two initials for the avatar.
	get initials() {
		const base = this.user?.name?.trim() || this.user?.email || '';
		const parts = base.split(/[\s@._-]+/).filter(Boolean);
		if (parts.length === 0) return '·';
		if (parts.length === 1) return parts[0].slice(0, 2).toUpperCase();
		return (parts[0][0] + parts[1][0]).toUpperCase();
	}

	clear() {
		this.user = null;
		this.csrf = '';
		// If you were logged in, an owner exists, so setup is not needed.
		this.needsSetup = false;
	}
}

export const session = new SessionStore();
