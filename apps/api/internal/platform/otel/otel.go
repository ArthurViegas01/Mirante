// Package otel centralizes OpenTelemetry setup. In F0 it installs a no-op
// provider so domain code can create spans freely; the real OTLP exporter is
// wired in F5 without touching call sites.
package otel

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

// Init installs the global tracer provider and returns a shutdown function.
// The serviceName/endpoint parameters are accepted now so the F5 swap to an
// OTLP exporter needs no signature change.
func Init(_ /*serviceName*/, _ /*endpoint*/ string) (trace.TracerProvider, func(context.Context) error) {
	tp := noop.NewTracerProvider()
	otel.SetTracerProvider(tp)
	return tp, func(context.Context) error { return nil }
}

// Tracer returns a named tracer from the global provider.
func Tracer(name string) trace.Tracer { return otel.Tracer(name) }
