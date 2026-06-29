// Package ratelimit is a small fixed-window limiter keyed by an arbitrary
// string (IP, email, route). It is used both for the login attempt cap and for
// per-IP HTTP throttling. Single-user app, single process: an in-memory map is
// sufficient.
package ratelimit

import (
	"sync"
	"time"
)

// defaultMaxKeys bounds how many distinct keys a limiter tracks at once, so
// attacker-controlled keys (forged IPs, arbitrary login e-mails) can't grow the
// map without limit (M2). On a 256MB VM this is comfortably small.
const defaultMaxKeys = 10000

// Limiter allows up to max hits per key within a rolling window.
type Limiter struct {
	mu      sync.Mutex
	max     int
	window  time.Duration
	hits    map[string]*counter
	maxKeys int
	now     func() time.Time
}

type counter struct {
	count   int
	resetAt time.Time
}

// New returns a limiter permitting max events per window per key.
func New(max int, window time.Duration) *Limiter {
	return &Limiter{
		max:     max,
		window:  window,
		hits:    make(map[string]*counter),
		maxKeys: defaultMaxKeys,
		now:     time.Now,
	}
}

// Allow records a hit and reports whether the key is still under the limit. The
// distinct-key count is hard-capped at maxKeys so attacker-controlled keys can't
// grow the map without bound (M2). When a new key arrives at the cap, expired
// counters are dropped first; if the map is still full, the entry nearest its
// reset is evicted to make room. Eviction never denies a legitimate caller — it
// only reclaims a counter that was about to reset anyway, so it can't be used to
// lock a real user out of login/signup.
func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	c, ok := l.hits[key]
	if ok && !now.After(c.resetAt) {
		if c.count >= l.max {
			return false
		}
		c.count++
		return true
	}

	// New key, or an existing key whose window has expired. Only a genuinely new
	// key grows the map, so the cap is enforced only there.
	if !ok && len(l.hits) >= l.maxKeys {
		l.reclaimSlot(now)
	}
	l.hits[key] = &counter{count: 1, resetAt: now.Add(l.window)}
	return true
}

// Reset clears a key's counter (e.g. after a successful login).
func (l *Limiter) Reset(key string) {
	l.mu.Lock()
	delete(l.hits, key)
	l.mu.Unlock()
}

// reclaimSlot frees at least one map slot in a single O(n) pass: it drops every
// expired counter and, if none were expired, evicts the one nearest its reset.
// Evicting the nearest-expiry counter (vs refusing the new key) keeps a flood of
// distinct attacker keys from locking a legitimate caller out of login/signup,
// while still bounding the map at maxKeys. Caller holds the lock.
func (l *Limiter) reclaimSlot(now time.Time) {
	var victim string
	var soonest time.Time
	haveVictim, removedAny := false, false
	for k, c := range l.hits {
		if now.After(c.resetAt) {
			delete(l.hits, k)
			removedAny = true
			continue
		}
		if !haveVictim || c.resetAt.Before(soonest) {
			soonest, victim, haveVictim = c.resetAt, k, true
		}
	}
	if !removedAny && haveVictim {
		delete(l.hits, victim)
	}
}
