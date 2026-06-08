// Package ratelimit is a small fixed-window limiter keyed by an arbitrary
// string (IP, email, route). It is used both for the login attempt cap and for
// per-IP HTTP throttling. Single-user app, single process: an in-memory map is
// sufficient.
package ratelimit

import (
	"sync"
	"time"
)

// Limiter allows up to max hits per key within a rolling window.
type Limiter struct {
	mu     sync.Mutex
	max    int
	window time.Duration
	hits   map[string]*counter
	now    func() time.Time
}

type counter struct {
	count   int
	resetAt time.Time
}

// New returns a limiter permitting max events per window per key.
func New(max int, window time.Duration) *Limiter {
	return &Limiter{
		max:    max,
		window: window,
		hits:   make(map[string]*counter),
		now:    time.Now,
	}
}

// Allow records a hit and reports whether the key is still under the limit.
func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	c, ok := l.hits[key]
	if !ok || now.After(c.resetAt) {
		l.hits[key] = &counter{count: 1, resetAt: now.Add(l.window)}
		l.gc(now)
		return true
	}
	if c.count >= l.max {
		return false
	}
	c.count++
	return true
}

// Reset clears a key's counter (e.g. after a successful login).
func (l *Limiter) Reset(key string) {
	l.mu.Lock()
	delete(l.hits, key)
	l.mu.Unlock()
}

// gc drops expired counters opportunistically. Caller holds the lock.
func (l *Limiter) gc(now time.Time) {
	if len(l.hits) < 1024 {
		return
	}
	for k, c := range l.hits {
		if now.After(c.resetAt) {
			delete(l.hits, k)
		}
	}
}
