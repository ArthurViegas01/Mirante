package db

import (
	"context"
	"database/sql/driver"
	"errors"
	"io"
	"strings"
	"time"
)

// Retry tuning. Four attempts with a short, capped exponential backoff add at
// most ~350ms to a doomed call while clearing the brief connectivity blips we see
// from Turso (a server-recycled stream, a momentary edge 502). A sustained Turso
// outage outlasts this — it is an upstream problem no client retry can mask.
const (
	maxAttempts = 4
	baseBackoff = 50 * time.Millisecond
	maxBackoff  = 1 * time.Second
)

// transientSubstrings are fragments of libSQL/network errors that mean the
// statement did not run (a dropped stream, an edge that could not reach the DB,
// a reset socket) and so are safe to retry on a fresh connection. Matched
// case-insensitively against the full (possibly wrapped) error string.
var transientSubstrings = []string{
	"stream is closed",    // server recycled the libSQL stream under us
	"bad connection",      // database/sql / driver bad-conn, as a string
	"connect to upstream", // Turso edge 502: gateway could not reach the DB
	"connection reset",    // peer reset mid-flight
	"broken pipe",         // write to a half-closed socket
	"unexpected eof",      // stream cut before a full response
	"i/o timeout",         // network stall
	"connection refused",  // Turso instance momentarily down (e.g. restart)
	"use of closed network connection",
}

// IsTransient reports whether err is a retryable connectivity failure rather than
// a real query/logic error. Caller cancellation (context.Canceled /
// DeadlineExceeded) is deliberately NOT transient: it is intentional, and
// retrying would fight the caller. driver.ErrBadConn and io.EOF are transient
// even when wrapped.
func IsTransient(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	if errors.Is(err, driver.ErrBadConn) || errors.Is(err, io.EOF) {
		return true
	}
	msg := strings.ToLower(err.Error())
	for _, s := range transientSubstrings {
		if strings.Contains(msg, s) {
			return true
		}
	}
	return false
}

// Retry runs fn, retrying on a transient connectivity error (see IsTransient)
// with a bounded, ctx-aware backoff. A non-transient error is returned at once;
// so is the last transient error once attempts are exhausted. fn must be safe to
// re-run — use it for reads and single idempotent writes, not for work whose
// commit may already have landed (see DB.WithTx for the transactional case).
func Retry(ctx context.Context, fn func() error) error {
	var err error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 && !backoff(ctx, attempt) {
			return ctx.Err()
		}
		if err = fn(); err == nil {
			return nil
		}
		if !IsTransient(err) {
			return err
		}
	}
	return err
}

// backoff waits for an exponentially growing, capped delay before attempt n
// (n >= 1), or returns false the moment ctx is cancelled.
func backoff(ctx context.Context, attempt int) bool {
	d := baseBackoff << (attempt - 1)
	if d > maxBackoff {
		d = maxBackoff
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-t.C:
		return true
	}
}
