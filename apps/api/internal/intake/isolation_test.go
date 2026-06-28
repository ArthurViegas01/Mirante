package intake

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/lumni/mirante/internal/platform/tenant"
)

func ctxFor(uid string) context.Context {
	return tenant.WithUserID(context.Background(), uid)
}

// TestUserIsolation is the security net for per-user scoping: dedup is per user
// (B staging the same digest is not deduped against A), and B can neither read nor
// mutate A's staged items.
func TestUserIsolation(t *testing.T) {
	svc := newService(t, nil, 60)
	ctxA := ctxFor("user-a")
	ctxB := ctxFor("user-b")
	raw := loadDigest(t)

	sumA, err := svc.Ingest(ctxA, [][]byte{raw})
	require.NoError(t, err)
	require.Greater(t, sumA.New, 0)

	// Same digest for B stages its own copies — the dedup key includes user_id.
	sumB, err := svc.Ingest(ctxB, [][]byte{raw})
	require.NoError(t, err)
	require.Equal(t, sumA.New, sumB.New)
	require.Equal(t, 0, sumB.Duplicate)

	listA, err := svc.List(ctxA, ListFilter{})
	require.NoError(t, err)
	listB, err := svc.List(ctxB, ListFilter{})
	require.NoError(t, err)
	require.Equal(t, len(listA), len(listB))

	// B cannot read or mutate A's item.
	aItem := listA[0]
	_, err = svc.Get(ctxB, aItem.ID)
	require.ErrorIs(t, err, ErrNotFound)
	require.ErrorIs(t, svc.Dismiss(ctxB, aItem.ID), ErrNotFound)

	// A's item is untouched by B's attempts.
	got, err := svc.Get(ctxA, aItem.ID)
	require.NoError(t, err)
	require.Equal(t, EstadoNovo, got.Estado)
}
