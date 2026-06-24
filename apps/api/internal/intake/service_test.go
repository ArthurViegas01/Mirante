package intake

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	idb "github.com/lumni/mirante/internal/platform/db"
	"github.com/lumni/mirante/internal/platform/migrate"
)

func newService(t *testing.T, skills SkillsProvider, minScore int) *Service {
	t.Helper()
	ctx := context.Background()
	database, err := idb.Open(ctx, ":memory:", "")
	require.NoError(t, err)
	t.Cleanup(func() { _ = database.Close() })
	require.NoError(t, migrate.Up(database.DB))
	return NewService(NewSQLiteRepo(database), skills, minScore)
}

func loadDigest(t *testing.T) []byte {
	t.Helper()
	raw, err := os.ReadFile("testdata/digest.eml")
	require.NoError(t, err)
	return raw
}

func TestIngestStagesAndDedupes(t *testing.T) {
	svc := newService(t, nil, 60)
	ctx := ctxFor("u1")
	raw := loadDigest(t)

	sum, err := svc.Ingest(ctx, [][]byte{raw})
	require.NoError(t, err)
	require.Equal(t, 1, sum.Emails)
	require.Greater(t, sum.New, 5)
	require.Equal(t, 0, sum.Duplicate)

	// Re-ingesting the same digest stages nothing new (dedup by source id).
	sum2, err := svc.Ingest(ctx, [][]byte{raw})
	require.NoError(t, err)
	require.Equal(t, 0, sum2.New)
	require.Equal(t, sum.New, sum2.Duplicate)

	// Listing returns them highest score first, all freshly staged.
	items, err := svc.List(ctx, ListFilter{})
	require.NoError(t, err)
	require.Len(t, items, sum.New)
	for i := range items {
		require.Equal(t, EstadoNovo, items[i].Estado)
		require.NotEmpty(t, items[i].FonteID)
		if i > 0 {
			require.GreaterOrEqual(t, items[i-1].Score, items[i].Score)
		}
	}
}

func TestIngestScoresWithSkills(t *testing.T) {
	skills := func(context.Context) ([]string, error) { return []string{"Shopify"}, nil }
	svc := newService(t, skills, 60)
	ctx := ctxFor("u1")

	_, err := svc.Ingest(ctx, [][]byte{loadDigest(t)})
	require.NoError(t, err)

	items, err := svc.List(ctx, ListFilter{})
	require.NoError(t, err)

	var shopify *Item
	for _, it := range items {
		if strings.Contains(strings.ToLower(it.Titulo), "shopify") {
			shopify = it
			break
		}
	}
	require.NotNil(t, shopify, "fixture should contain a Shopify project")
	require.Contains(t, shopify.Skills, "Shopify")
}

func TestIngestSkipsUnparseable(t *testing.T) {
	svc := newService(t, nil, 60)
	sum, err := svc.Ingest(ctxFor("u1"), [][]byte{[]byte("isto não é um e-mail")})
	require.NoError(t, err)
	require.Equal(t, 1, sum.Failed)
	require.Equal(t, 0, sum.New)
	require.Equal(t, 0, sum.Emails)
}

func TestDismiss(t *testing.T) {
	svc := newService(t, nil, 60)
	ctx := ctxFor("u1")
	_, err := svc.Ingest(ctx, [][]byte{loadDigest(t)})
	require.NoError(t, err)

	items, err := svc.List(ctx, ListFilter{Estado: EstadoNovo})
	require.NoError(t, err)
	require.NotEmpty(t, items)
	target := items[0]

	require.NoError(t, svc.Dismiss(ctx, target.ID))

	got, err := svc.Get(ctx, target.ID)
	require.NoError(t, err)
	require.Equal(t, EstadoDescartado, got.Estado)

	novo, err := svc.List(ctx, ListFilter{Estado: EstadoNovo})
	require.NoError(t, err)
	for _, it := range novo {
		require.NotEqual(t, target.ID, it.ID)
	}
}

func TestListMinScoreFilter(t *testing.T) {
	svc := newService(t, nil, 0)
	ctx := ctxFor("u1")
	_, err := svc.Ingest(ctx, [][]byte{loadDigest(t)})
	require.NoError(t, err)

	all, err := svc.List(ctx, ListFilter{})
	require.NoError(t, err)
	require.NotEmpty(t, all)

	// A high floor keeps only the strongest leads — a strict subset (the
	// 73-proposal lead is always filtered out).
	top, err := svc.List(ctx, ListFilter{MinScore: 60})
	require.NoError(t, err)
	require.Less(t, len(top), len(all))
	for _, it := range top {
		require.GreaterOrEqual(t, it.Score, 60)
	}
}
