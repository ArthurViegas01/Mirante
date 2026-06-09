package llm

import (
	"sync"
	"time"
)

type window struct {
	start time.Time
	count int
}

type fixedWindowLimiter struct {
	mu        sync.Mutex
	perMinute int
	windows   map[string]*window
	now       func() time.Time
}

// NewRouteLimiter caps each route to perMinute calls within a one-minute fixed
// window. perMinute <= 0 disables limiting (Allow always returns true).
func NewRouteLimiter(perMinute int) RouteLimiter {
	return &fixedWindowLimiter{
		perMinute: perMinute,
		windows:   map[string]*window{},
		now:       time.Now,
	}
}

func (l *fixedWindowLimiter) Allow(route string) bool {
	if l.perMinute <= 0 {
		return true
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	w := l.windows[route]
	if w == nil || now.Sub(w.start) >= time.Minute {
		l.windows[route] = &window{start: now, count: 1}
		return true
	}
	if w.count >= l.perMinute {
		return false
	}
	w.count++
	return true
}
