package cv

import (
	"context"
	"strings"
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

func TestProfileEmptyByDefault(t *testing.T) {
	p, err := newService(t).GetProfile(context.Background())
	require.NoError(t, err)
	require.Equal(t, "", p.Titulo)
	require.Equal(t, "", p.TituloAlvo)
}

func TestProfileUpsert(t *testing.T) {
	ctx := context.Background()
	svc := newService(t)

	p, err := svc.SaveProfile(ctx, ProfileInput{Titulo: "Dev Backend", TituloAlvo: "Staff Engineer"})
	require.NoError(t, err)
	require.Equal(t, "Dev Backend", p.Titulo)
	require.Equal(t, "Staff Engineer", p.TituloAlvo)

	got, err := svc.GetProfile(ctx)
	require.NoError(t, err)
	require.Equal(t, "Staff Engineer", got.TituloAlvo)

	// A second save overwrites the same singleton row.
	_, err = svc.SaveProfile(ctx, ProfileInput{Titulo: "Dev Pleno"})
	require.NoError(t, err)
	got2, _ := svc.GetProfile(ctx)
	require.Equal(t, "Dev Pleno", got2.Titulo)
	require.Equal(t, "", got2.TituloAlvo) // cleared
}

func TestProfileValidation(t *testing.T) {
	_, err := newService(t).SaveProfile(context.Background(), ProfileInput{Titulo: strings.Repeat("x", 121)})
	require.ErrorIs(t, err, ErrInvalid)
}

func TestProfileSkills(t *testing.T) {
	ctx := context.Background()
	svc := newService(t)

	// golang→Go (canonical), "react"/"React" dedup, blank skipped, unknown kept.
	p, err := svc.SaveProfile(ctx, ProfileInput{
		Titulo: "Dev",
		Skills: []string{"golang", "React", "react", "Salesforce", "  "},
	})
	require.NoError(t, err)
	require.Equal(t, []string{"Go", "React", "Salesforce"}, p.Skills) // returned sorted

	got, err := svc.GetProfile(ctx)
	require.NoError(t, err)
	require.Equal(t, []string{"Go", "React", "Salesforce"}, got.Skills)

	// Re-saving replaces the whole set.
	p2, err := svc.SaveProfile(ctx, ProfileInput{Titulo: "Dev", Skills: []string{"Python"}})
	require.NoError(t, err)
	require.Equal(t, []string{"Python"}, p2.Skills)
}
