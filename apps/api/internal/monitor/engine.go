package monitor

import (
	"context"
	"log/slog"
	"time"
)

// EventSink receives live events for the SSE stream. It is implemented by the
// SSE hub (platform/sse) and wired in cmd/server; monitor never imports the hub.
type EventSink interface {
	Emit(id int64, eventType string, data []byte)
}

// AlertChannel is a pluggable external alert delivery (email/webhook). v1 ships
// none — the in-app alert (a persisted row + SSE event) is the only delivery.
type AlertChannel interface {
	Name() string
	Send(ctx context.Context, a Alert) error
}

// Notifier fans a transition alert out to external channels, isolating errors
// and bounding each call with its own timeout. With zero channels it is a no-op.
type Notifier struct {
	channels []AlertChannel
	log      *slog.Logger
}

// NewNotifier builds a notifier over zero or more channels.
func NewNotifier(log *slog.Logger, channels ...AlertChannel) *Notifier {
	return &Notifier{channels: channels, log: log}
}

// Dispatch delivers the alert to every external channel.
func (n *Notifier) Dispatch(ctx context.Context, a Alert) {
	for _, ch := range n.channels {
		cctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		if err := ch.Send(cctx, a); err != nil {
			n.log.Warn("alert channel failed", "channel", ch.Name(), "err", err)
		}
		cancel()
	}
}

// Engine runs one check end to end: probe → derive → persist → emit.
type Engine struct {
	repo     Repository
	checker  Checker
	notifier *Notifier
	sink     EventSink
	log      *slog.Logger
}

// NewEngine builds the check engine.
func NewEngine(repo Repository, checker Checker, notifier *Notifier, sink EventSink, log *slog.Logger) *Engine {
	return &Engine{repo: repo, checker: checker, notifier: notifier, sink: sink, log: log}
}

// RunCheck probes svc, derives the new state, persists the result and any
// transition atomically, then emits the live event and dispatches the alert. It
// mutates svc's status/counters in place for the caller's next tick.
func (e *Engine) RunCheck(ctx context.Context, svc *Service) error {
	sample := e.checker.Check(ctx, svc)
	res := Derive(DeriveInput{
		Prev:            svc.CurrentStatus,
		ConsecFailures:  svc.ConsecutiveFailures,
		ConsecSuccesses: svc.ConsecutiveSuccesses,
		Sample:          sample,
		T:               Thresholds{DegradedMs: svc.DegradedThresholdMs, N: svc.AntiFlapN, K: svc.RecoveryK},
	})

	out, err := e.repo.RecordCheck(ctx, RecordCheckInput{
		Service:    svc,
		From:       svc.CurrentStatus,
		Result:     res,
		LatencyMs:  sample.LatencyMs,
		StatusCode: sample.StatusCode,
		ErrorKind:  errorKind(sample),
		CheckedAt:  time.Now().UTC(),
	})
	if err != nil {
		return err
	}

	svc.CurrentStatus = res.State
	svc.ConsecutiveFailures = res.ConsecFailures
	svc.ConsecutiveSuccesses = res.ConsecSuccesses

	if out.Event != nil && e.sink != nil {
		e.sink.Emit(out.Event.ID, out.Event.Type, out.Event.Data)
	}
	if out.Alert != nil && e.notifier != nil {
		e.notifier.Dispatch(ctx, *out.Alert)
	}
	return nil
}

func errorKind(s Sample) string {
	switch {
	case !s.Responded:
		return "no_response"
	case !s.OK:
		return "wrong_status"
	default:
		return ""
	}
}
