package projects

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	idb "github.com/lumni/mirante/internal/platform/db"
	"github.com/lumni/mirante/internal/platform/migrate"
)

func newService(t *testing.T) *Service {
	t.Helper()
	ctx := context.Background()
	database, err := idb.Open(ctx, ":memory:", "")
	require.NoError(t, err)
	t.Cleanup(func() { _ = database.Close() })
	require.NoError(t, migrate.Up(database.DB))
	return NewService(NewSQLiteRepo(database))
}

func TestCreateAndGet(t *testing.T) {
	ctx := context.Background()
	svc := newService(t)

	p, err := svc.Create(ctx, CreateInput{Nome: "Mirante", Codinome: "mirante", Tags: []string{"Go", "SvelteKit"}})
	require.NoError(t, err)
	require.NotEmpty(t, p.ID)
	require.Equal(t, StatusIdeia, p.Status)
	require.Equal(t, VisPessoal, p.Visibilidade)
	require.ElementsMatch(t, []string{"Go", "SvelteKit"}, p.Tags)

	got, err := svc.Get(ctx, p.ID)
	require.NoError(t, err)
	require.Equal(t, "Mirante", got.Nome)
}

func TestCreateValidation(t *testing.T) {
	ctx := context.Background()
	svc := newService(t)

	_, err := svc.Create(ctx, CreateInput{Nome: "   "})
	require.ErrorIs(t, err, ErrInvalid)

	_, err = svc.Create(ctx, CreateInput{Nome: "X", Repo: "not-a-url"})
	require.ErrorIs(t, err, ErrInvalid)
}

func TestListFilterByStatus(t *testing.T) {
	ctx := context.Background()
	svc := newService(t)
	_, _ = svc.Create(ctx, CreateInput{Nome: "A", Status: StatusAtivo})
	_, _ = svc.Create(ctx, CreateInput{Nome: "B", Status: StatusIdeia})

	ativos, err := svc.List(ctx, ListFilter{Status: string(StatusAtivo)})
	require.NoError(t, err)
	require.Len(t, ativos, 1)
	require.Equal(t, "A", ativos[0].Nome)

	all, err := svc.List(ctx, ListFilter{})
	require.NoError(t, err)
	require.Len(t, all, 2)
}

func TestUpdate(t *testing.T) {
	ctx := context.Background()
	svc := newService(t)
	p, _ := svc.Create(ctx, CreateInput{Nome: "Old"})

	newName := "New"
	st := StatusAtivo
	tags := []string{"x"}
	up, err := svc.Update(ctx, p.ID, UpdateInput{Nome: &newName, Status: &st, Tags: &tags})
	require.NoError(t, err)
	require.Equal(t, "New", up.Nome)
	require.Equal(t, StatusAtivo, up.Status)
	require.Equal(t, []string{"x"}, up.Tags)
}

func TestDelete(t *testing.T) {
	ctx := context.Background()
	svc := newService(t)
	p, _ := svc.Create(ctx, CreateInput{Nome: "Temp"})

	require.NoError(t, svc.Delete(ctx, p.ID))
	_, err := svc.Get(ctx, p.ID)
	require.ErrorIs(t, err, ErrNotFound)
	require.ErrorIs(t, svc.Delete(ctx, p.ID), ErrNotFound)
}

func TestLinks(t *testing.T) {
	ctx := context.Background()
	svc := newService(t)
	p, _ := svc.Create(ctx, CreateInput{Nome: "WithLinks"})

	up, err := svc.AddLink(ctx, p.ID, LinkInput{Label: "Prod", URL: "https://mirante.app", Kind: "prod"})
	require.NoError(t, err)
	require.Len(t, up.Links, 1)
	linkID := up.Links[0].ID

	_, err = svc.AddLink(ctx, p.ID, LinkInput{Label: "Bad", URL: "nope"})
	require.ErrorIs(t, err, ErrInvalid)

	require.NoError(t, svc.RemoveLink(ctx, p.ID, linkID))
	got, _ := svc.Get(ctx, p.ID)
	require.Len(t, got.Links, 0)
}

func TestGetNotFound(t *testing.T) {
	svc := newService(t)
	_, err := svc.Get(context.Background(), "nonexistent")
	require.ErrorIs(t, err, ErrNotFound)
}
