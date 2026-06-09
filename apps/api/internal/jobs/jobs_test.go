package jobs

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/lumni/mirante/internal/llm"
	idb "github.com/lumni/mirante/internal/platform/db"
	"github.com/lumni/mirante/internal/platform/httpx"
	"github.com/lumni/mirante/internal/platform/migrate"
)

func newService(t *testing.T, client *llm.Client) *Service {
	return newServiceFetch(t, client, nil)
}

func newServiceFetch(t *testing.T, client *llm.Client, fetcher *httpx.Fetcher) *Service {
	t.Helper()
	ctx := context.Background()
	database, err := idb.Open(ctx, ":memory:", "")
	require.NoError(t, err)
	t.Cleanup(func() { _ = database.Close() })
	require.NoError(t, migrate.Up(database.DB))
	return NewService(NewSQLiteRepo(database), client, fetcher)
}

// localFetcher allows private IPs so tests can hit an httptest server on loopback.
func localFetcher() *httpx.Fetcher {
	return httpx.NewFetcher(httpx.Policy{AllowPrivateIPs: true, MaxBodyBytes: 1 << 20})
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

const jsonLDPage = `<html><head>
<script type="application/ld+json">
{"@context":"https://schema.org/","@type":"JobPosting","title":"Engenheiro de Software",
"description":"<p>Trabalhar com <strong>Go</strong>, React e PostgreSQL.</p>",
"hiringOrganization":{"@type":"Organization","name":"Acme"},
"jobLocation":{"@type":"Place","address":{"@type":"PostalAddress","addressLocality":"Porto Alegre","addressRegion":"RS"}},
"jobLocationType":"TELECOMMUTE"}
</script></head><body>página com ruído</body></html>`

func htmlServer(t *testing.T, body string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestImportDraftJSONLD(t *testing.T) {
	srv := htmlServer(t, jsonLDPage)
	svc := newServiceFetch(t, nil, localFetcher())

	d, err := svc.ImportDraft(context.Background(), srv.URL)
	require.NoError(t, err)
	require.Equal(t, "json-ld", d.Fonte)
	require.Equal(t, "Engenheiro de Software", d.Titulo)
	require.Equal(t, "Acme", d.Empresa)
	require.Equal(t, ModeloRemoto, d.Modelo)
	require.Equal(t, "Porto Alegre, RS", d.Localizacao)
	require.Contains(t, d.Descricao, "Go")
	require.ElementsMatch(t, []string{"Go", "React", "PostgreSQL"}, d.Skills)
}

func TestImportDraftLLMFallback(t *testing.T) {
	srv := htmlServer(t, `<html><body><h1>Vaga</h1><p>Backend com Go e Docker</p></body></html>`)
	client := llm.NewClient(llm.NewMock(
		`{"titulo":"Dev Backend","empresa":"Beta","descricao":"Backend com Go e Docker","localizacao":"Remoto","modelo":"remoto","senioridade":"pleno"}`,
	), nil, nil)
	svc := newServiceFetch(t, client, localFetcher())

	d, err := svc.ImportDraft(context.Background(), srv.URL)
	require.NoError(t, err)
	require.Equal(t, "llm", d.Fonte)
	require.Equal(t, "Dev Backend", d.Titulo)
	require.Equal(t, ModeloRemoto, d.Modelo)
	require.ElementsMatch(t, []string{"Go", "Docker"}, d.Skills)
}

func TestImportFailedNoSource(t *testing.T) {
	srv := htmlServer(t, `<html><body><p>página sem dados estruturados</p></body></html>`)
	svc := newServiceFetch(t, nil, localFetcher()) // no LLM, no JSON-LD
	_, err := svc.ImportDraft(context.Background(), srv.URL)
	require.ErrorIs(t, err, ErrImportFailed)
}

func TestImportUnavailableAndBadURL(t *testing.T) {
	noFetcher := newService(t, nil)
	_, err := noFetcher.ImportDraft(context.Background(), "https://example.com/job")
	require.ErrorIs(t, err, ErrImportUnavailable)

	withFetcher := newServiceFetch(t, nil, localFetcher())
	_, err = withFetcher.ImportDraft(context.Background(), "not-a-url")
	require.ErrorIs(t, err, ErrInvalid)
}
