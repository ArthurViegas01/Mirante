package applications

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

func TestCreateAndDefaults(t *testing.T) {
	ctx := context.Background()
	svc := newService(t)
	a, err := svc.Create(ctx, CreateInput{JobID: "job1", Titulo: "Backend Eng", Empresa: "Acme"})
	require.NoError(t, err)
	require.NotEmpty(t, a.ID)
	require.Equal(t, StatusInteresse, a.Status)
	require.Equal(t, "job1", a.JobID)
	require.Equal(t, "Acme", a.Empresa)
}

func TestCreateValidation(t *testing.T) {
	ctx := context.Background()
	svc := newService(t)
	_, err := svc.Create(ctx, CreateInput{Titulo: "  "})
	require.ErrorIs(t, err, ErrInvalid)
	_, err = svc.Create(ctx, CreateInput{Titulo: "X", Status: "foo"})
	require.ErrorIs(t, err, ErrInvalid)
	_, err = svc.Create(ctx, CreateInput{Titulo: "X", DataAcao: "amanhã"})
	require.ErrorIs(t, err, ErrInvalid)
}

func TestListByStatus(t *testing.T) {
	ctx := context.Background()
	svc := newService(t)
	_, _ = svc.Create(ctx, CreateInput{Titulo: "A", Status: StatusAplicado})
	_, _ = svc.Create(ctx, CreateInput{Titulo: "B", Status: StatusEntrevista})

	aplicado, err := svc.List(ctx, ListFilter{Status: string(StatusAplicado)})
	require.NoError(t, err)
	require.Len(t, aplicado, 1)
	require.Equal(t, "A", aplicado[0].Titulo)

	all, err := svc.List(ctx, ListFilter{})
	require.NoError(t, err)
	require.Len(t, all, 2)
}

func TestUpdateAndDelete(t *testing.T) {
	ctx := context.Background()
	svc := newService(t)
	a, _ := svc.Create(ctx, CreateInput{Titulo: "X"})

	st := StatusEntrevista
	prox := "Enviar follow-up"
	data := "2026-07-01"
	up, err := svc.Update(ctx, a.ID, UpdateInput{Status: &st, ProximaAcao: &prox, DataAcao: &data})
	require.NoError(t, err)
	require.Equal(t, StatusEntrevista, up.Status)
	require.Equal(t, "Enviar follow-up", up.ProximaAcao)
	require.Equal(t, "2026-07-01", up.DataAcao)

	require.NoError(t, svc.Delete(ctx, a.ID))
	_, err = svc.Get(ctx, a.ID)
	require.ErrorIs(t, err, ErrNotFound)
	require.ErrorIs(t, svc.Delete(ctx, a.ID), ErrNotFound)
}
