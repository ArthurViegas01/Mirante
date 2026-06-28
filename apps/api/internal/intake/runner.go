package intake

import (
	"context"
	"log/slog"
	"time"

	"github.com/lumni/mirante/internal/platform/tenant"
)

// MessageSource yields raw RFC 822 e-mails to ingest. The IMAP poller is the
// production implementation; tests use a fake. Keeping it an interface lets the
// orchestration (Runner) be tested without a live mail server.
type MessageSource interface {
	Fetch(ctx context.Context) ([][]byte, error)
}

// OwnerResolver returns the user id the polled inbox belongs to (the instance
// admin). It returns "" with no error when no owner exists yet, so the poller idles
// until the instance is claimed via first-run signup.
type OwnerResolver func(ctx context.Context) (string, error)

// Runner periodically pulls e-mails from a MessageSource and ingests them under the
// owner's identity. A single in-process worker (ADR-0002, single instance).
type Runner struct {
	source   MessageSource
	svc      *Service
	owner    OwnerResolver
	interval time.Duration
	log      *slog.Logger
}

// NewRunner builds the intake poller.
func NewRunner(source MessageSource, svc *Service, owner OwnerResolver, interval time.Duration, log *slog.Logger) *Runner {
	if interval <= 0 {
		interval = 15 * time.Minute
	}
	return &Runner{source: source, svc: svc, owner: owner, interval: interval, log: log}
}

// Start runs the poll loop until ctx is cancelled; it returns immediately. The
// returned channel closes once the loop has fully stopped, for graceful shutdown.
func (r *Runner) Start(ctx context.Context) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(done)
		r.runOnce(ctx) // poll once on boot, then on each tick
		ticker := time.NewTicker(r.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				r.runOnce(ctx)
			}
		}
	}()
	return done
}

// runOnce performs a single poll+ingest, logging the outcome. Errors are logged,
// never fatal — the next tick retries.
func (r *Runner) runOnce(ctx context.Context) {
	ownerID, err := r.owner(ctx)
	if err != nil {
		r.log.Warn("intake: resolve owner", "err", err)
		return
	}
	if ownerID == "" {
		return // instance not claimed yet
	}
	raws, err := r.source.Fetch(ctx)
	if err != nil {
		r.log.Warn("intake: fetch", "err", err)
		return
	}
	if len(raws) == 0 {
		return
	}
	sum, err := r.svc.Ingest(tenant.WithUserID(ctx, ownerID), raws)
	if err != nil {
		r.log.Warn("intake: ingest", "err", err)
		return
	}
	if sum.New > 0 || sum.Failed > 0 {
		r.log.Info("intake poll",
			"emails", sum.Emails, "new", sum.New, "duplicate", sum.Duplicate, "failed", sum.Failed)
	}
}
