package monitor

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/lumni/mirante/internal/platform/tenant"
)

func ctxFor(uid string) context.Context {
	return tenant.WithUserID(context.Background(), uid)
}

// TestUserIsolation: a user only ever sees and controls their own services. The
// per-user SSE fan-out is covered separately in platform/sse.
func TestUserIsolation(t *testing.T) {
	mgr := NewManager(NewSQLiteRepo(openTestDB(t)))
	ctxA := ctxFor("user-a")
	ctxB := ctxFor("user-b")

	input := func(nome string) CreateServiceInput {
		return CreateServiceInput{
			ProjectID: "proj1", Nome: nome, Kind: KindHTTP, Target: "https://app.example.test",
		}
	}

	sa, err := mgr.CreateService(ctxA, input("A's service"))
	require.NoError(t, err)
	// Same project id is fine for a different user (the per-project limit is scoped).
	sb, err := mgr.CreateService(ctxB, input("B's service"))
	require.NoError(t, err)

	listA, err := mgr.ListServices(ctxA, "")
	require.NoError(t, err)
	require.Len(t, listA, 1)
	require.Equal(t, sa.ID, listA[0].ID)

	listB, err := mgr.ListServices(ctxB, "proj1")
	require.NoError(t, err)
	require.Len(t, listB, 1)
	require.Equal(t, sb.ID, listB[0].ID)

	// B cannot read, detail, update, toggle, or delete A's service.
	_, err = mgr.GetService(ctxB, sa.ID)
	require.ErrorIs(t, err, ErrNotFound)
	_, err = mgr.Detail(ctxB, sa.ID)
	require.ErrorIs(t, err, ErrNotFound)
	nome := "hijacked"
	_, err = mgr.UpdateService(ctxB, sa.ID, UpdateServiceInput{Nome: &nome})
	require.ErrorIs(t, err, ErrNotFound)
	_, err = mgr.SetEnabled(ctxB, sa.ID, false)
	require.ErrorIs(t, err, ErrNotFound)
	require.ErrorIs(t, mgr.DeleteService(ctxB, sa.ID), ErrNotFound)

	// A's service is untouched.
	got, err := mgr.GetService(ctxA, sa.ID)
	require.NoError(t, err)
	require.Equal(t, "A's service", got.Nome)
	require.True(t, got.Enabled)
}
