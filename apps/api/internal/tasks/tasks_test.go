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

func TestCreateAndGet(t *testing.T) {
	ctx := context.Background()
	svc, _ := newService(t)

	tk, err := svc.Create(ctx, CreateInput{Titulo: "Escrever ADR", Tags: []string{"docs", "F2"}})
	require.NoError(t, err)
	require.NotEmpty(t, tk.ID)
	require.Equal(t, StatusAFazer, tk.Status)
	require.Equal(t, PrioridadeMedia, tk.Prioridade)
	require.ElementsMatch(t, []string{"docs", "F2"}, tk.Tags)

	got, err := svc.Get(ctx, tk.ID)
	require.NoError(t, err)
	require.Equal(t, "Escrever ADR", got.Titulo)
}

func TestCreateValidation(t *testing.T) {
	ctx := context.Background()
	svc, _ := newService(t)

	_, err := svc.Create(ctx, CreateInput{Titulo: "   "})
	require.ErrorIs(t, err, ErrInvalid)

	_, err = svc.Create(ctx, CreateInput{Titulo: "X", Prazo: "amanhã"})
	require.ErrorIs(t, err, ErrInvalid)

	_, err = svc.Create(ctx, CreateInput{Titulo: "X", Prioridade: "urgentíssima"})
	require.ErrorIs(t, err, ErrInvalid)
}

func TestListFilters(t *testing.T) {
	ctx := context.Background()
	svc, database := newService(t)
	// project_id is FK-checked, so the linked projects must exist first.
	_, err := database.ExecContext(ctx, `INSERT INTO projects (id, nome) VALUES ('p1', 'P1'), ('p2', 'P2')`)
	require.NoError(t, err)
	_, _ = svc.Create(ctx, CreateInput{Titulo: "A", Status: StatusFazendo, ProjectID: "p1"})
	_, _ = svc.Create(ctx, CreateInput{Titulo: "B", Status: StatusAFazer, ProjectID: "p2"})
	_, _ = svc.Create(ctx, CreateInput{Titulo: "C", Status: StatusAFazer, ProjectID: "p1"})

	fazendo, err := svc.List(ctx, ListFilter{Status: string(StatusFazendo)})
	require.NoError(t, err)
	require.Len(t, fazendo, 1)
	require.Equal(t, "A", fazendo[0].Titulo)

	p1, err := svc.List(ctx, ListFilter{ProjectID: "p1"})
	require.NoError(t, err)
	require.Len(t, p1, 2)

	p1Todo, err := svc.List(ctx, ListFilter{ProjectID: "p1", Status: string(StatusAFazer)})
	require.NoError(t, err)
	require.Len(t, p1Todo, 1)
	require.Equal(t, "C", p1Todo[0].Titulo)

	all, err := svc.List(ctx, ListFilter{})
	require.NoError(t, err)
	require.Len(t, all, 3)
}

func TestUpdate(t *testing.T) {
	ctx := context.Background()
	svc, _ := newService(t)
	tk, _ := svc.Create(ctx, CreateInput{Titulo: "Old"})

	newTitle := "New"
	st := StatusFeito
	prio := PrioridadeAlta
	prazo := "2026-07-01"
	tags := []string{"x"}
	up, err := svc.Update(ctx, tk.ID, UpdateInput{
		Titulo: &newTitle, Status: &st, Prioridade: &prio, Prazo: &prazo, Tags: &tags,
	})
	require.NoError(t, err)
	require.Equal(t, "New", up.Titulo)
	require.Equal(t, StatusFeito, up.Status)
	require.Equal(t, PrioridadeAlta, up.Prioridade)
	require.Equal(t, "2026-07-01", up.Prazo)
	require.Equal(t, []string{"x"}, up.Tags)
}

func TestDelete(t *testing.T) {
	ctx := context.Background()
	svc, _ := newService(t)
	tk, _ := svc.Create(ctx, CreateInput{Titulo: "Temp"})

	require.NoError(t, svc.Delete(ctx, tk.ID))
	_, err := svc.Get(ctx, tk.ID)
	require.ErrorIs(t, err, ErrNotFound)
	require.ErrorIs(t, svc.Delete(ctx, tk.ID), ErrNotFound)
}

func TestTagsRoundTrip(t *testing.T) {
	ctx := context.Background()
	svc, _ := newService(t)
	tk, _ := svc.Create(ctx, CreateInput{Titulo: "WithTags", Tags: []string{"a", "b"}})

	empty := []string{}
	up, err := svc.Update(ctx, tk.ID, UpdateInput{Tags: &empty})
	require.NoError(t, err)
	require.Len(t, up.Tags, 0)
}

func TestProjectUnlinkOnDelete(t *testing.T) {
	ctx := context.Background()
	svc, database := newService(t)

	// A project owns a task; deleting the project must unlink (SET NULL), not
	// delete, the task — finished work outlives the project.
	_, err := database.ExecContext(ctx, `INSERT INTO projects (id, nome) VALUES (?, ?)`, "proj1", "Mirante")
	require.NoError(t, err)

	tk, err := svc.Create(ctx, CreateInput{Titulo: "Linked", ProjectID: "proj1"})
	require.NoError(t, err)
	require.Equal(t, "proj1", tk.ProjectID)

	_, err = database.ExecContext(ctx, `DELETE FROM projects WHERE id = ?`, "proj1")
	require.NoError(t, err)

	got, err := svc.Get(ctx, tk.ID)
	require.NoError(t, err)
	require.Empty(t, got.ProjectID)
}

func TestGetNotFound(t *testing.T) {
	svc, _ := newService(t)
	_, err := svc.Get(context.Background(), "nonexistent")
	require.ErrorIs(t, err, ErrNotFound)
}
