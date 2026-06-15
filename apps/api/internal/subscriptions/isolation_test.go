package subscriptions

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
// see or touch user A's subscriptions, even when both attach costs to the same
// project (isolation is by owner, not by project).
func TestUserIsolation(t *testing.T) {
	svc, _ := newService(t)
	ctxA := ctxFor("user-a")
	ctxB := ctxFor("user-b")

	// Both users add a subscription to the same project p1.
	sa, err := svc.Create(ctxA, CreateInput{ProjectID: "p1", Nome: "A's cost", ValorCents: 1900, Moeda: MoedaUSD})
	require.NoError(t, err)
	sb, err := svc.Create(ctxB, CreateInput{ProjectID: "p1", Nome: "B's cost", ValorCents: 4990, Moeda: MoedaBRL})
	require.NoError(t, err)

	// Each user lists only their own subscription, even filtered by project.
	listA, err := svc.List(ctxA, ListFilter{})
	require.NoError(t, err)
	require.Len(t, listA, 1)
	require.Equal(t, sa.ID, listA[0].ID)

	listB, err := svc.List(ctxB, ListFilter{ProjectID: "p1"})
	require.NoError(t, err)
	require.Len(t, listB, 1)
	require.Equal(t, sb.ID, listB[0].ID)

	// B cannot read, update, or delete A's subscription.
	_, err = svc.Get(ctxB, sa.ID)
	require.ErrorIs(t, err, ErrNotFound)
	hijack := "hijacked"
	_, err = svc.Update(ctxB, sa.ID, UpdateInput{Nome: &hijack})
	require.ErrorIs(t, err, ErrNotFound)
	require.ErrorIs(t, svc.Delete(ctxB, sa.ID), ErrNotFound)

	// A's subscription is untouched by B's attempts.
	got, err := svc.Get(ctxA, sa.ID)
	require.NoError(t, err)
	require.Equal(t, "A's cost", got.Nome)
	require.Equal(t, 1900, got.ValorCents)
}
