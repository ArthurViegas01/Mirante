package tasks

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
// see or touch user A's tasks or tags, and the same tags may coexist across
// users without bleeding.
func TestUserIsolation(t *testing.T) {
	svc, _ := newService(t)
	ctxA := ctxFor("user-a")
	ctxB := ctxFor("user-b")

	ta, err := svc.Create(ctxA, CreateInput{Titulo: "A's task", Tags: []string{"go"}})
	require.NoError(t, err)

	// Same tag name is allowed for a different user (tags are per user).
	tb, err := svc.Create(ctxB, CreateInput{Titulo: "B's task", Tags: []string{"rust"}})
	require.NoError(t, err)

	// Each user lists only their own task.
	listA, err := svc.List(ctxA, ListFilter{})
	require.NoError(t, err)
	require.Len(t, listA, 1)
	require.Equal(t, ta.ID, listA[0].ID)
	require.Equal(t, []string{"go"}, listA[0].Tags)

	listB, err := svc.List(ctxB, ListFilter{})
	require.NoError(t, err)
	require.Len(t, listB, 1)
	require.Equal(t, tb.ID, listB[0].ID)
	require.Equal(t, []string{"rust"}, listB[0].Tags) // tags don't bleed across users

	// B cannot read, update, or delete A's task.
	_, err = svc.Get(ctxB, ta.ID)
	require.ErrorIs(t, err, ErrNotFound)
	hijack := "hijacked"
	_, err = svc.Update(ctxB, ta.ID, UpdateInput{Titulo: &hijack})
	require.ErrorIs(t, err, ErrNotFound)
	require.ErrorIs(t, svc.Delete(ctxB, ta.ID), ErrNotFound)

	// A's task is untouched by B's attempts.
	got, err := svc.Get(ctxA, ta.ID)
	require.NoError(t, err)
	require.Equal(t, "A's task", got.Titulo)
	require.Equal(t, []string{"go"}, got.Tags)
}
