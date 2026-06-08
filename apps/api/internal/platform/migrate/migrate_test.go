package migrate

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/lumni/mirante/internal/platform/db"
)

func TestUpCreatesAndDownRemoves(t *testing.T) {
	ctx := context.Background()
	database, err := db.Open(ctx, ":memory:", "")
	require.NoError(t, err)
	t.Cleanup(func() { _ = database.Close() })

	require.NoError(t, Up(database.DB))
	require.True(t, tableExists(t, database.DB, "users"))
	require.True(t, tableExists(t, database.DB, "sessions"))

	require.NoError(t, Reset(database.DB))
	require.False(t, tableExists(t, database.DB, "users"))
	require.False(t, tableExists(t, database.DB, "sessions"))
}

func tableExists(t *testing.T, sqlDB *sql.DB, name string) bool {
	t.Helper()
	var n int
	err := sqlDB.QueryRow(
		`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?`, name).Scan(&n)
	require.NoError(t, err)
	return n > 0
}
