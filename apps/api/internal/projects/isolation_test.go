package projects

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
// see or touch user A's projects, links or tags, and the same codinome/tags may
// coexist across users.
func TestUserIsolation(t *testing.T) {
	svc := newService(t)
	ctxA := ctxFor("user-a")
	ctxB := ctxFor("user-b")

	pa, err := svc.Create(ctxA, CreateInput{Nome: "A's project", Codinome: "alpha", Tags: []string{"go"}})
	require.NoError(t, err)
	_, err = svc.AddLink(ctxA, pa.ID, LinkInput{Label: "Prod", URL: "https://a.example", Kind: "prod"})
	require.NoError(t, err)

	// Same codinome is allowed for a different user (uniqueness is per user).
	pb, err := svc.Create(ctxB, CreateInput{Nome: "B's project", Codinome: "alpha", Tags: []string{"rust"}})
	require.NoError(t, err)

	// Each user lists only their own project.
	listA, err := svc.List(ctxA, ListFilter{})
	require.NoError(t, err)
	require.Len(t, listA, 1)
	require.Equal(t, pa.ID, listA[0].ID)

	listB, err := svc.List(ctxB, ListFilter{})
	require.NoError(t, err)
	require.Len(t, listB, 1)
	require.Equal(t, pb.ID, listB[0].ID)
	require.Equal(t, []string{"rust"}, listB[0].Tags) // tags don't bleed across users

	// B cannot read, update, delete, or link A's project.
	_, err = svc.Get(ctxB, pa.ID)
	require.ErrorIs(t, err, ErrNotFound)
	hijack := "hijacked"
	_, err = svc.Update(ctxB, pa.ID, UpdateInput{Nome: &hijack})
	require.ErrorIs(t, err, ErrNotFound)
	require.ErrorIs(t, svc.Delete(ctxB, pa.ID), ErrNotFound)
	_, err = svc.AddLink(ctxB, pa.ID, LinkInput{Label: "x", URL: "https://x.example"})
	require.ErrorIs(t, err, ErrNotFound)

	// A's project is untouched by B's attempts.
	got, err := svc.Get(ctxA, pa.ID)
	require.NoError(t, err)
	require.Equal(t, "A's project", got.Nome)
	require.Len(t, got.Links, 1)
	require.Equal(t, []string{"go"}, got.Tags)
}
