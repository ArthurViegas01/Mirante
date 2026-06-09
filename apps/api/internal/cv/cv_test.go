package cv

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/lumni/mirante/internal/llm"
	idb "github.com/lumni/mirante/internal/platform/db"
	"github.com/lumni/mirante/internal/platform/migrate"
)

func newService(t *testing.T) *Service { return newServiceLLM(t, nil) }

func newServiceLLM(t *testing.T, client *llm.Client) *Service {
	t.Helper()
	ctx := context.Background()
	database, err := idb.Open(ctx, ":memory:", "")
	require.NoError(t, err)
	t.Cleanup(func() { _ = database.Close() })
	require.NoError(t, migrate.Up(database.DB))
	return NewService(NewSQLiteRepo(database), client)
}

func TestProfileEmptyByDefault(t *testing.T) {
	p, err := newService(t).GetProfile(context.Background())
	require.NoError(t, err)
	require.Equal(t, "", p.Titulo)
	require.Empty(t, p.Skills)
	require.Empty(t, p.Experiences)
	require.Empty(t, p.Education)
}

func TestSaveCVFull(t *testing.T) {
	ctx := context.Background()
	svc := newService(t)

	p, err := svc.SaveCV(ctx, CVInput{
		Nome: "Arthur", Titulo: "Dev Backend", TituloAlvo: "Staff Engineer",
		Skills: []string{"golang", "React", "react"}, // → Go, React (canonical, deduped)
		Experiences: []ExperienceInput{
			{Empresa: "Acme", Cargo: "Backend", Inicio: "2022", Fim: "atual", Descricao: "Go e PostgreSQL"},
			{Empresa: "", Cargo: "", Descricao: ""}, // blank row → skipped
		},
		Education: []EducationInput{
			{Instituicao: "UFRGS", Curso: "Ciência da Computação", Inicio: "2016", Fim: "2021"},
		},
	})
	require.NoError(t, err)
	require.Equal(t, []string{"Go", "React"}, p.Skills)
	require.Len(t, p.Experiences, 1)
	require.Equal(t, "Acme", p.Experiences[0].Empresa)
	require.NotEmpty(t, p.Experiences[0].ID)
	require.Len(t, p.Education, 1)
	require.Equal(t, "UFRGS", p.Education[0].Instituicao)

	got, err := svc.GetProfile(ctx)
	require.NoError(t, err)
	require.Equal(t, "Staff Engineer", got.TituloAlvo)
	require.Len(t, got.Experiences, 1)
	require.Len(t, got.Education, 1)
}

func TestSaveProfilePreservesSkillsAndCV(t *testing.T) {
	ctx := context.Background()
	svc := newService(t)

	_, err := svc.SaveCV(ctx, CVInput{
		Titulo: "Dev", Skills: []string{"Go", "Docker"},
		Experiences: []ExperienceInput{{Empresa: "Acme", Cargo: "Eng"}},
	})
	require.NoError(t, err)

	// Partial identity update (no skills) must preserve skills AND experiences —
	// this is the header quick-edit path that previously wiped skills.
	p, err := svc.SaveProfile(ctx, ProfileInput{Titulo: "Senior Dev"})
	require.NoError(t, err)
	require.Equal(t, "Senior Dev", p.Titulo)
	require.Equal(t, []string{"Docker", "Go"}, p.Skills) // sorted, preserved
	require.Len(t, p.Experiences, 1)

	// An explicit skills pointer DOES replace them.
	replacement := []string{"Python"}
	p2, err := svc.SaveProfile(ctx, ProfileInput{Titulo: "Senior Dev", Skills: &replacement})
	require.NoError(t, err)
	require.Equal(t, []string{"Python"}, p2.Skills)
}

func TestProfileValidation(t *testing.T) {
	_, err := newService(t).SaveCV(context.Background(), CVInput{Titulo: strings.Repeat("x", 121)})
	require.ErrorIs(t, err, ErrInvalid)
}

func TestImportDraft(t *testing.T) {
	ctx := context.Background()
	client := llm.NewClient(llm.NewMock(
		`{"nome":"Arthur","titulo":"Full-Stack Engineer","titulo_alvo":"Staff Engineer",
		  "resumo":"Engenheiro de software.","skills":["golang","FastAPI","react","react"],
		  "experiences":[{"empresa":"Dataglass","cargo":"Dev","inicio":"2023","fim":"2024","descricao":"Django + React"},
		                 {"empresa":"","cargo":"","descricao":""}],
		  "education":[{"instituicao":"PUCRS","curso":"Eng. de Software","inicio":"2021","fim":"2025"}]}`,
	), nil, nil)
	svc := newServiceLLM(t, client)

	d, err := svc.ImportDraft(ctx, ImportInput{Text: "currículo colado…"})
	require.NoError(t, err)
	require.Equal(t, "Arthur", d.Nome)
	require.Equal(t, "Staff Engineer", d.TituloAlvo)
	require.Contains(t, d.Skills, "Go") // golang → Go (canonical)
	require.Contains(t, d.Skills, "React")
	require.Len(t, d.Experiences, 1) // blank row dropped
	require.Equal(t, "Dataglass", d.Experiences[0].Empresa)
	require.Len(t, d.Education, 1)

	// The draft is NOT persisted until saved.
	got, _ := svc.GetProfile(ctx)
	require.Empty(t, got.Experiences)
}

func TestImportRequiresLLMAndText(t *testing.T) {
	ctx := context.Background()
	_, err := newService(t).ImportDraft(ctx, ImportInput{Text: "x"})
	require.ErrorIs(t, err, ErrLLMUnavailable) // nil client

	withLLM := newServiceLLM(t, llm.NewClient(llm.NewMock("{}"), nil, nil))
	_, err = withLLM.ImportDraft(ctx, ImportInput{Text: "   "})
	require.ErrorIs(t, err, ErrInvalid) // empty text checked first
}
