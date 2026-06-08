// Package logging builds the application's structured logger and redacts
// sensitive attributes so secrets never reach logs (or, later, OTel exporters).
package logging

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

// New returns a structured logger: JSON in production, text in development.
func New(appEnv string) *slog.Logger {
	opts := &slog.HandlerOptions{Level: slog.LevelInfo}
	var base slog.Handler
	if appEnv == "production" {
		base = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		base = slog.NewTextHandler(os.Stdout, opts)
	}
	return slog.New(&redactHandler{Handler: base})
}

// sensitiveKeys are masked wherever they appear as an attribute key.
var sensitiveKeys = map[string]bool{
	"password":      true,
	"password_hash": true,
	"token":         true,
	"token_hash":    true,
	"authorization": true,
	"secret":        true,
	"secret_key":    true,
	"api_key":       true,
	"apikey":        true,
	"auth_token":    true,
	"dsn":           true,
	"cookie":        true,
}

type redactHandler struct {
	slog.Handler
}

func (h *redactHandler) Handle(ctx context.Context, r slog.Record) error {
	nr := slog.NewRecord(r.Time, r.Level, r.Message, r.PC)
	r.Attrs(func(a slog.Attr) bool {
		nr.AddAttrs(redact(a))
		return true
	})
	return h.Handler.Handle(ctx, nr)
}

func (h *redactHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	out := make([]slog.Attr, len(attrs))
	for i, a := range attrs {
		out[i] = redact(a)
	}
	return &redactHandler{Handler: h.Handler.WithAttrs(out)}
}

func (h *redactHandler) WithGroup(name string) slog.Handler {
	return &redactHandler{Handler: h.Handler.WithGroup(name)}
}

func redact(a slog.Attr) slog.Attr {
	if sensitiveKeys[strings.ToLower(a.Key)] {
		return slog.String(a.Key, "[REDACTED]")
	}
	return a
}
