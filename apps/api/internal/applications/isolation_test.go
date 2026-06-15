package applications

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/lumni/mirante/internal/platform/tenant"
)

func ctxFor(uid string) context.Context {
	return tenant.WithUserID(context.Background(), uid)
}

// TestUserIsolation is the security net for per-user scoping: user B must never
// see or touch user A's candidaturas, and identical applications may coexist
// across users.
func TestUserIsolation(t *testing.T) {
	svc := newService(t)
	ctxA := ctxFor("user-a")
	ctxB := ctxFor("user-b")

	aa, err := svc.Create(ctxA, CreateInput{Titulo: "A's candidatura", Empresa: "Acme", Status: StatusAplicado})
	require.NoError(t, err)

	// The same candidatura is allowed for a different user (scoping is per user).
	ab, err := svc.Create(ctxB, CreateInput{Titulo: "B's candidatura", Empresa: "Acme", Status: StatusAplicado})
	require.NoError(t, err)

	// Each user lists only their own candidaturas, including filtered listings.
	listA, err := svc.List(ctxA, ListFilter{})
	require.NoError(t, err)
	require.Len(t, listA, 1)
	require.Equal(t, aa.ID, listA[0].ID)

	listB, err := svc.List(ctxB, ListFilter{})
	require.NoError(t, err)
	require.Len(t, listB, 1)
	require.Equal(t, ab.ID, listB[0].ID)

	// A status filter must not leak the other user's rows either.
	filteredB, err := svc.List(ctxB, ListFilter{Status: string(StatusAplicado)})
	require.NoError(t, err)
	require.Len(t, filteredB, 1)
	require.Equal(t, ab.ID, filteredB[0].ID)

	// B cannot read, update, or delete A's candidatura.
	_, err = svc.Get(ctxB, aa.ID)
	require.ErrorIs(t, err, ErrNotFound)
	hijack := "hijacked"
	_, err = svc.Update(ctxB, aa.ID, UpdateInput{Titulo: &hijack})
	require.ErrorIs(t, err, ErrNotFound)
	require.ErrorIs(t, svc.Delete(ctxB, aa.ID), ErrNotFound)

	// A's candidatura is untouched by B's attempts.
	got, err := svc.Get(ctxA, aa.ID)
	require.NoError(t, err)
	require.Equal(t, "A's candidatura", got.Titulo)
}
