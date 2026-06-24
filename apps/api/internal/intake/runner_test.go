package intake

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type fakeSource struct {
	raws  [][]byte
	err   error
	calls int
}

func (f *fakeSource) Fetch(context.Context) ([][]byte, error) {
	f.calls++
	return f.raws, f.err
}

func quietLog() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }

func TestRunnerIngestsUnderOwner(t *testing.T) {
	svc := newService(t, nil, 60)
	src := &fakeSource{raws: [][]byte{loadDigest(t)}}
	owner := func(context.Context) (string, error) { return "owner-1", nil }
	r := NewRunner(src, svc, owner, time.Minute, quietLog())

	r.runOnce(context.Background())

	// Items landed under the resolved owner, and only there.
	items, err := svc.List(ctxFor("owner-1"), ListFilter{})
	require.NoError(t, err)
	require.NotEmpty(t, items)

	other, err := svc.List(ctxFor("someone-else"), ListFilter{})
	require.NoError(t, err)
	require.Empty(t, other)
}

func TestRunnerSkipsWhenNoOwner(t *testing.T) {
	svc := newService(t, nil, 60)
	src := &fakeSource{raws: [][]byte{loadDigest(t)}}
	owner := func(context.Context) (string, error) { return "", nil } // unclaimed instance
	r := NewRunner(src, svc, owner, time.Minute, quietLog())

	r.runOnce(context.Background())
	require.Equal(t, 0, src.calls) // never fetched — nothing to attribute items to
}
