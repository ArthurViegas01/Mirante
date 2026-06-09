package jobs

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/lumni/mirante/internal/llm"
	idb "github.com/lumni/mirante/internal/platform/db"
	"github.com/lumni/mirante/internal/platform/migrate"
)

func newService(t *testing.T, client *llm.Client) *Service {
	t.Helper()
	ctx := context.Background()
	database, err := idb.Open(ctx, ":memory:", "")
	require.NoError(t, err)
	t.Cleanup(func() { _ = database.Close() })
	require.NoError(t, migrate.Up(database.DB))
	return NewService(NewSQLiteRepo(database), client)
}

func TestCreateExtractsSkills(t *testing.T) {
	ctx := context.Background()
	svc := newService(t, nil)

	j, err := svc.Create(ctx, CreateInput{
		Titulo:    "Dev Backend",
		Descricao: "Buscamos pessoa dev com Go, React e Docker (bônus: Kubernetes).",
	})
	require.NoError(t, err)
	require.Equal(t, ModeloIndefinido, j.Modelo)
	require.ElementsMatch(t, []string{"Go", "React", "Docker", "Kubernetes"}, j.Skills)
}

func TestCreateValidation(t *testing.T) {
	ctx := context.Background()
	svc := newService(t, nil)

	_, err := svc.Create(ctx, CreateInput{Titulo: "   "})
	require.ErrorIs(t, err, ErrInvalid)

	_, err = svc.Create(ctx, CreateInput{Titulo: "X", URL: "not-a-url"})
	require.ErrorIs(t, err, ErrInvalid)

	_, err = svc.Create(ctx, CreateInput{Titulo: "X", Modelo: "fulltime"})
	require.ErrorIs(t, err, ErrInvalid)
}

func TestUpdateReextractsSkills(t *testing.T) {
	ctx := context.Background()
	svc := newService(t, nil)
	j, _ := svc.Create(ctx, CreateInput{Titulo: "V", Descricao: "Go e Docker"})
	require.ElementsMatch(t, []string{"Go", "Docker"}, j.Skills)

	desc := "Agora é Python com Django e PostgreSQL"
	up, err := svc.Update(ctx, j.ID, UpdateInput{Descricao: &desc})
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"Python", "Django", "PostgreSQL"}, up.Skills)
}

func TestDeleteCascadesSkills(t *testing.T) {
	ctx := context.Background()
	svc := newService(t, nil)
	j, _ := svc.Create(ctx, CreateInput{Titulo: "V", Descricao: "Go"})

	require.NoError(t, svc.Delete(ctx, j.ID))
	_, err := svc.Get(ctx, j.ID)
	require.ErrorIs(t, err, ErrNotFound)
	require.ErrorIs(t, svc.Delete(ctx, j.ID), ErrNotFound)
}

func TestEnrichUnavailable(t *testing.T) {
	ctx := context.Background()
	svc := newService(t, nil) // no LLM client
	j, _ := svc.Create(ctx, CreateInput{Titulo: "V", Descricao: "Go"})
	_, err := svc.Enrich(ctx, j.ID)
	require.ErrorIs(t, err, ErrLLMUnavailable)
}

func TestEnrichWithMock(t *testing.T) {
	ctx := context.Background()
	client := llm.NewClient(
		llm.NewMock(`{"empresa":"ACME","senioridade":"pleno","modelo":"remoto","resumo":"Vaga de backend Go."}`),
		nil, nil,
	)
	svc := newService(t, client)

	j, _ := svc.Create(ctx, CreateInput{Titulo: "Backend", Descricao: "Vaga com Go e Docker."})
	require.Equal(t, "", j.Empresa)
	require.Equal(t, ModeloIndefinido, j.Modelo)

	enriched, err := svc.Enrich(ctx, j.ID)
	require.NoError(t, err)
	require.Equal(t, "ACME", enriched.Empresa)
	require.Equal(t, "pleno", enriched.Senioridade)
	require.Equal(t, ModeloRemoto, enriched.Modelo)
	require.Equal(t, "Vaga de backend Go.", enriched.Resumo)
	// Deterministic skills are untouched by enrichment.
	require.ElementsMatch(t, []string{"Go", "Docker"}, enriched.Skills)
}

func TestList(t *testing.T) {
	ctx := context.Background()
	svc := newService(t, nil)
	_, _ = svc.Create(ctx, CreateInput{Titulo: "A"})
	_, _ = svc.Create(ctx, CreateInput{Titulo: "B"})
	all, err := svc.List(ctx)
	require.NoError(t, err)
	require.Len(t, all, 2)
}
