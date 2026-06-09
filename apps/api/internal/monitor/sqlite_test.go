package monitor

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	idb "github.com/lumni/mirante/internal/platform/db"
)

func insertCheck(t *testing.T, d *idb.DB, sid string, at time.Time, outcome Status, latencyMs int) {
	t.Helper()
	_, err := d.ExecContext(context.Background(),
		`INSERT INTO check_results (service_id, checked_at, ok, outcome, latency_ms) VALUES (?, ?, ?, ?, ?)`,
		sid, idb.FormatTime(at), boolToInt(outcome != StatusDown), string(outcome), latencyMs)
	require.NoError(t, err)
}

// TestCompactRollupAndPrune verifies that compaction rolls raw checks older than
// the cutoff into hourly buckets, prunes them, and that uptime over a window
// spanning both rollups and surviving raw rows is unchanged (no loss, no double
// counting) — plus idempotency and the on-conflict merge path.
func TestCompactRollupAndPrune(t *testing.T) {
	ctx := context.Background()
	database := openTestDB(t)
	repo := NewSQLiteRepo(database)
	require.NoError(t, repo.CreateService(ctx, newService("svc1")))

	base := time.Now().UTC()
	old := base.Add(-20 * 24 * time.Hour) // older than 14d retention → compacted
	recent := base.Add(-2 * time.Hour)    // within retention → kept raw

	// Three old checks in one hour (up, up, down) and two recent ups.
	insertCheck(t, database, "svc1", old, StatusUp, 100)
	insertCheck(t, database, "svc1", old, StatusUp, 200)
	insertCheck(t, database, "svc1", old, StatusDown, 300)
	insertCheck(t, database, "svc1", recent, StatusUp, 50)
	insertCheck(t, database, "svc1", recent, StatusUp, 50)

	rawCount := func() int {
		var n int
		require.NoError(t, database.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM check_results WHERE service_id = ?`, "svc1").Scan(&n))
		return n
	}
	rollup := func() (rows, samples, ups, sumLat int) {
		require.NoError(t, database.QueryRowContext(ctx,
			`SELECT COUNT(*), COALESCE(SUM(samples),0), COALESCE(SUM(ups),0), COALESCE(SUM(sum_latency_ms),0)
			 FROM check_rollups WHERE service_id = ?`, "svc1").Scan(&rows, &samples, &ups, &sumLat))
		return
	}

	// Baseline: all five samples are raw; 30d uptime = 4/5.
	before30, err := repo.Uptime(ctx, "svc1", 24*30)
	require.NoError(t, err)
	require.Equal(t, 5, before30.Samples)
	require.InDelta(t, 0.8, before30.UpRatio, 1e-9)

	// Compact everything older than 14d (cutoff floored to the hour).
	cutoff := base.Add(-14 * 24 * time.Hour).Truncate(time.Hour)
	pruned, err := repo.Compact(ctx, cutoff)
	require.NoError(t, err)
	require.Equal(t, 3, pruned)

	require.Equal(t, 2, rawCount()) // only the two recent rows survive
	rows, samples, ups, sumLat := rollup()
	require.Equal(t, 1, rows) // one hourly bucket
	require.Equal(t, 3, samples)
	require.Equal(t, 2, ups)      // up + up, the down excluded
	require.Equal(t, 600, sumLat) // 100 + 200 + 300

	// 30d uptime is identical post-compaction: rollup (3/2) + raw (2/2) = 5 / 4.
	after30, err := repo.Uptime(ctx, "svc1", 24*30)
	require.NoError(t, err)
	require.Equal(t, 5, after30.Samples)
	require.InDelta(t, 0.8, after30.UpRatio, 1e-9)

	// 24h uptime sees only the recent raw rows; the old bucket is before the
	// window start hour and is excluded → 2 / 2.
	day, err := repo.Uptime(ctx, "svc1", 24)
	require.NoError(t, err)
	require.Equal(t, 2, day.Samples)
	require.InDelta(t, 1.0, day.UpRatio, 1e-9)

	// Idempotent: re-running with the same cutoff prunes nothing and does not
	// double the rollup.
	again, err := repo.Compact(ctx, cutoff)
	require.NoError(t, err)
	require.Equal(t, 0, again)
	_, samples, _, _ = rollup()
	require.Equal(t, 3, samples)

	// On-conflict merge: a late row in an already-compacted hour sums into the
	// existing bucket rather than creating a duplicate.
	insertCheck(t, database, "svc1", old, StatusDown, 0)
	merged, err := repo.Compact(ctx, cutoff)
	require.NoError(t, err)
	require.Equal(t, 1, merged)
	rows, samples, ups, _ = rollup()
	require.Equal(t, 1, rows)
	require.Equal(t, 4, samples)
	require.Equal(t, 2, ups)
}
