// Package otel centralizes OpenTelemetry setup. With no OTLP endpoint configured
// it installs a no-op provider (spans are free but not exported), so dev and
// tests pay nothing; setting OTEL_EXPORTER_OTLP_ENDPOINT swaps in a real
// OTLP/HTTP trace exporter without touching any call site.
package otel

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

// Init installs the global tracer provider + W3C propagator and returns a
// shutdown function (flushes pending spans). serviceName tags the resource; an
// empty endpoint keeps the no-op provider. The exporter reads the standard
// OTEL_EXPORTER_OTLP_* environment for endpoint/headers/TLS, so it stays spec
// compliant; a failure to build it degrades to no-op rather than blocking boot.
func Init(serviceName, endpoint string) (trace.TracerProvider, func(context.Context) error) {
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{}))

	if endpoint == "" {
		return installNoop()
	}

	exp, err := otlptracehttp.New(context.Background())
	if err != nil {
		slog.Default().Warn("otel exporter init failed; tracing disabled", "err", err)
		return installNoop()
	}

	// Empty schema URL merges cleanly with the SDK's default resource (which
	// carries its own schema); avoids pinning a semconv version here.
	res, err := resource.Merge(resource.Default(),
		resource.NewWithAttributes("", attribute.String("service.name", serviceName)))
	if err != nil {
		res = resource.Default()
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	slog.Default().Info("otel tracing enabled", "endpoint", endpoint, "service", serviceName)
	return tp, tp.Shutdown
}

func installNoop() (trace.TracerProvider, func(context.Context) error) {
	tp := noop.NewTracerProvider()
	otel.SetTracerProvider(tp)
	return tp, func(context.Context) error { return nil }
}

// Tracer returns a named tracer from the global provider.
func Tracer(name string) trace.Tracer { return otel.Tracer(name) }
