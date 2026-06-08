package tasks

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	idb "github.com/lumni/mirante/internal/platform/db"
	"github.com/lumni/mirante/internal/platform/migrate"
)

func newService(t *testing.T) (*Service, *idb.DB) {
	t.Helper()
	ctx := context.Background()
	database, err := idb.Open(ctx, ":memory:", "")
	require.NoError(t, err)
	t.Cleanup(func() { _ = database.Close() })
	require.NoError(t, migrate.Up(database.DB))
	return NewService(NewSQLiteRepo(database)), database
}

func ptr(s string) *string { return &s }

func insertProject(t *testing.T, db *idb.DB, id string) {
	t.Helper()
	_, err := db.ExecContext(context.Background(),
		`INSERT INTO projects (id, nome) VALUES (?, ?)`, id, "Proj "+id)
	require.NoError(t, err)
}

func TestCreateAndGetDefaults(t *testing.T) {
	ctx := context.Background()
	svc, _ := newService(t)

	tk, err := svc.Create(ctx, CreateInput{Titulo: "Escrever testes", Tags: []string{"go", "f2"}})
	require.NoError(t, err)
	require.NotEmpty(t, tk.ID)
	require.Equal(t, StatusAFazer, tk.Status)
	require.Equal(t, PrioMedia, tk.Prioridade)
	require.Nil(t, tk.Prazo)
	require.Nil(t, tk.ProjectID)
	require.Nil(t, tk.JobID)
	require.ElementsMatch(t, []string{"go", "f2"}, tk.Tags)

	got, err := svc.Get(ctx, tk.ID)
	require.NoError(t, err)
	require.Equal(t, "Escrever testes", got.Titulo)
}

func TestCreateValidation(t *testing.T) {
	ctx := context.Background()
	svc, _ := newService(t)

	_, err := svc.Create(ctx, CreateInput{Titulo: "   "})
	require.ErrorIs(t, err, ErrInvalid)

	_, err = svc.Create(ctx, CreateInput{Titulo: "x", Prioridade: Prioridade("urgente")})
	require.ErrorIs(t, err, ErrInvalid)

	_, err = svc.Create(ctx, CreateInput{Titulo: "x", Status: Status("done")})
	require.ErrorIs(t, err, ErrInvalid)

	_, err = svc.Create(ctx, CreateInput{Titulo: "x", Prazo: ptr("01/07/2026")})
	require.ErrorIs(t, err, ErrInvalid)
}

func TestPrazoSetAndClear(t *testing.T) {
	ctx := context.Background()
	svc, _ := newService(t)

	tk, err := svc.Create(ctx, CreateInput{Titulo: "com prazo", Prazo: ptr(" 2026-07-01 ")})
	require.NoError(t, err)
	require.NotNil(t, tk.Prazo)
	require.Equal(t, "2026-07-01", *tk.Prazo)

	// A non-nil empty string clears the deadline to NULL.
	up, err := svc.Update(ctx, tk.ID, UpdateInput{Prazo: ptr("")})
	require.NoError(t, err)
	require.Nil(t, up.Prazo)
}

func TestListFilters(t *testing.T) {
	ctx := context.Background()
	svc, db := newService(t)
	insertProject(t, db, "p1")
	insertProject(t, db, "p2")

	_, _ = svc.Create(ctx, CreateInput{Titulo: "A", Status: StatusFazendo, Prioridade: PrioAlta, ProjectID: ptr("p1")})
	_, _ = svc.Create(ctx, CreateInput{Titulo: "B", Status: StatusAFazer, ProjectID: ptr("p2")})
	_, _ = svc.Create(ctx, CreateInput{Titulo: "C", Status: StatusFeito})

	fazendo, err := svc.List(ctx, ListFilter{Status: string(StatusFazendo)})
	require.NoError(t, err)
	require.Len(t, fazendo, 1)
	require.Equal(t, "A", fazendo[0].Titulo)

	p2, err := svc.List(ctx, ListFilter{ProjectID: "p2"})
	require.NoError(t, err)
	require.Len(t, p2, 1)
	require.Equal(t, "B", p2[0].Titulo)

	alta, err := svc.List(ctx, ListFilter{Prioridade: string(PrioAlta)})
	require.NoError(t, err)
	require.Len(t, alta, 1)

	all, err := svc.List(ctx, ListFilter{})
	require.NoError(t, err)
	require.Len(t, all, 3)
}

func TestUpdateStatusAndTags(t *testing.T) {
	ctx := context.Background()
	svc, _ := newService(t)
	tk, _ := svc.Create(ctx, CreateInput{Titulo: "mover"})

	st := StatusFeito
	tags := []string{"done"}
	up, err := svc.Update(ctx, tk.ID, UpdateInput{Status: &st, Tags: &tags})
	require.NoError(t, err)
	require.Equal(t, StatusFeito, up.Status)
	require.Equal(t, []string{"done"}, up.Tags)

	// Clearing tags with an empty slice removes them.
	empty := []string{}
	up, err = svc.Update(ctx, tk.ID, UpdateInput{Tags: &empty})
	require.NoError(t, err)
	require.Empty(t, up.Tags)
}

func TestProjectDeleteDetachesTask(t *testing.T) {
	ctx := context.Background()
	svc, db := newService(t)
	insertProject(t, db, "p1")

	tk, err := svc.Create(ctx, CreateInput{Titulo: "linked", ProjectID: ptr("p1")})
	require.NoError(t, err)
	require.NotNil(t, tk.ProjectID)
	require.Equal(t, "p1", *tk.ProjectID)

	_, err = db.ExecContext(ctx, `DELETE FROM projects WHERE id = ?`, "p1")
	require.NoError(t, err)

	// ON DELETE SET NULL: the task survives, just unlinked.
	got, err := svc.Get(ctx, tk.ID)
	require.NoError(t, err)
	require.Nil(t, got.ProjectID)
}

func TestDelete(t *testing.T) {
	ctx := context.Background()
	svc, _ := newService(t)
	tk, _ := svc.Create(ctx, CreateInput{Titulo: "temp"})

	require.NoError(t, svc.Delete(ctx, tk.ID))
	_, err := svc.Get(ctx, tk.ID)
	require.ErrorIs(t, err, ErrNotFound)
	require.ErrorIs(t, svc.Delete(ctx, tk.ID), ErrNotFound)
}

func TestGetNotFound(t *testing.T) {
	svc, _ := newService(t)
	_, err := svc.Get(context.Background(), "nonexistent")
	require.ErrorIs(t, err, ErrNotFound)
}
