package otel

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

func TestInitNoopWithoutEndpoint(t *testing.T) {
	tp, shutdown := Init("mirante-test", "")
	require.NotNil(t, tp)
	require.IsType(t, noop.NewTracerProvider(), tp)
	require.NoError(t, shutdown(context.Background()))

	// Tracer is usable and never nil, even with the no-op provider.
	_, span := Tracer("test").Start(context.Background(), "op")
	span.End()
}

func TestInitRealProviderWithEndpoint(t *testing.T) {
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4318")
	tp, shutdown := Init("mirante-test", "http://localhost:4318")
	require.NotNil(t, tp)
	require.IsType(t, (*sdktrace.TracerProvider)(nil), tp) // real SDK provider, not no-op

	_, span := Tracer("test").Start(context.Background(), "op")
	span.End()

	// Flush; ignore any transport error to an absent collector (no spans queued).
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = shutdown(ctx)
}
