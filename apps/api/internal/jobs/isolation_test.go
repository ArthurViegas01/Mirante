package jobs

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
// see or touch user A's jobs, and a job's extracted skills must not bleed across
// users.
func TestUserIsolation(t *testing.T) {
	svc := newService(t, nil)
	ctxA := ctxFor("user-a")
	ctxB := ctxFor("user-b")

	ja, err := svc.Create(ctxA, CreateInput{Titulo: "A's job", Descricao: "Buscamos dev com Go e Docker."})
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"Go", "Docker"}, ja.Skills)

	jb, err := svc.Create(ctxB, CreateInput{Titulo: "B's job", Descricao: "Vaga com Python e Django."})
	require.NoError(t, err)

	// Each user lists only their own job.
	listA, err := svc.List(ctxA)
	require.NoError(t, err)
	require.Len(t, listA, 1)
	require.Equal(t, ja.ID, listA[0].ID)

	listB, err := svc.List(ctxB)
	require.NoError(t, err)
	require.Len(t, listB, 1)
	require.Equal(t, jb.ID, listB[0].ID)
	require.ElementsMatch(t, []string{"Python", "Django"}, listB[0].Skills) // skills don't bleed across users

	// B cannot read, update, or delete A's job.
	_, err = svc.Get(ctxB, ja.ID)
	require.ErrorIs(t, err, ErrNotFound)
	hijack := "hijacked"
	_, err = svc.Update(ctxB, ja.ID, UpdateInput{Titulo: &hijack})
	require.ErrorIs(t, err, ErrNotFound)
	require.ErrorIs(t, svc.Delete(ctxB, ja.ID), ErrNotFound)

	// A's job is untouched by B's attempts.
	got, err := svc.Get(ctxA, ja.ID)
	require.NoError(t, err)
	require.Equal(t, "A's job", got.Titulo)
	require.ElementsMatch(t, []string{"Go", "Docker"}, got.Skills)
}
