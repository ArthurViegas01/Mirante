package subscriptions

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
	// project_id is FK-checked; seed a couple of projects.
	_, err = database.ExecContext(ctx, `INSERT INTO projects (id, nome) VALUES ('p1', 'P1'), ('p2', 'P2')`)
	require.NoError(t, err)
	return NewService(NewSQLiteRepo(database)), database
}

func TestCreateAndGetDefaults(t *testing.T) {
	ctx := context.Background()
	svc, _ := newService(t)

	sub, err := svc.Create(ctx, CreateInput{ProjectID: "p1", Nome: "Netlify Pro", ValorCents: 1900, Moeda: MoedaUSD})
	require.NoError(t, err)
	require.NotEmpty(t, sub.ID)
	require.Equal(t, CicloMensal, sub.Ciclo)
	require.Equal(t, MoedaUSD, sub.Moeda)
	require.True(t, sub.Ativo)

	// Moeda defaults to BRL when omitted.
	br, err := svc.Create(ctx, CreateInput{ProjectID: "p1", Nome: "Domínio"})
	require.NoError(t, err)
	require.Equal(t, MoedaBRL, br.Moeda)

	got, err := svc.Get(ctx, sub.ID)
	require.NoError(t, err)
	require.Equal(t, "Netlify Pro", got.Nome)
	require.Equal(t, 1900, got.ValorCents)
}

func TestCreateValidation(t *testing.T) {
	ctx := context.Background()
	svc, _ := newService(t)

	_, err := svc.Create(ctx, CreateInput{ProjectID: "p1", Nome: "   "})
	require.ErrorIs(t, err, ErrInvalid)

	_, err = svc.Create(ctx, CreateInput{ProjectID: "p1", Nome: "X", ValorCents: -1})
	require.ErrorIs(t, err, ErrInvalid)

	_, err = svc.Create(ctx, CreateInput{ProjectID: "p1", Nome: "X", Moeda: "EUR"})
	require.ErrorIs(t, err, ErrInvalid)

	_, err = svc.Create(ctx, CreateInput{ProjectID: "p1", Nome: "X", Ciclo: "semanal"})
	require.ErrorIs(t, err, ErrInvalid)

	_, err = svc.Create(ctx, CreateInput{ProjectID: "", Nome: "X"})
	require.ErrorIs(t, err, ErrInvalid)
}

func TestListByProject(t *testing.T) {
	ctx := context.Background()
	svc, _ := newService(t)
	_, _ = svc.Create(ctx, CreateInput{ProjectID: "p1", Nome: "A"})
	_, _ = svc.Create(ctx, CreateInput{ProjectID: "p1", Nome: "B"})
	_, _ = svc.Create(ctx, CreateInput{ProjectID: "p2", Nome: "C"})

	p1, err := svc.List(ctx, ListFilter{ProjectID: "p1"})
	require.NoError(t, err)
	require.Len(t, p1, 2)

	all, err := svc.List(ctx, ListFilter{})
	require.NoError(t, err)
	require.Len(t, all, 3)
}

func TestUpdate(t *testing.T) {
	ctx := context.Background()
	svc, _ := newService(t)
	sub, _ := svc.Create(ctx, CreateInput{ProjectID: "p1", Nome: "Old", ValorCents: 100})

	nome := "New"
	valor := 4990
	ciclo := CicloAnual
	ativo := false
	svcID := "svc-123"
	up, err := svc.Update(ctx, sub.ID, UpdateInput{
		Nome: &nome, ValorCents: &valor, Ciclo: &ciclo, Ativo: &ativo, ServiceID: &svcID,
	})
	require.NoError(t, err)
	require.Equal(t, "New", up.Nome)
	require.Equal(t, 4990, up.ValorCents)
	require.Equal(t, CicloAnual, up.Ciclo)
	require.False(t, up.Ativo)
	require.Equal(t, "svc-123", up.ServiceID)
}

func TestDelete(t *testing.T) {
	ctx := context.Background()
	svc, _ := newService(t)
	sub, _ := svc.Create(ctx, CreateInput{ProjectID: "p1", Nome: "Temp"})

	require.NoError(t, svc.Delete(ctx, sub.ID))
	_, err := svc.Get(ctx, sub.ID)
	require.ErrorIs(t, err, ErrNotFound)
	require.ErrorIs(t, svc.Delete(ctx, sub.ID), ErrNotFound)
}

func TestCascadeOnProjectDelete(t *testing.T) {
	ctx := context.Background()
	svc, database := newService(t)
	sub, err := svc.Create(ctx, CreateInput{ProjectID: "p1", Nome: "Some cost"})
	require.NoError(t, err)

	_, err = database.ExecContext(ctx, `DELETE FROM projects WHERE id = ?`, "p1")
	require.NoError(t, err)

	_, err = svc.Get(ctx, sub.ID)
	require.ErrorIs(t, err, ErrNotFound) // CASCADE removed the subscription
}
