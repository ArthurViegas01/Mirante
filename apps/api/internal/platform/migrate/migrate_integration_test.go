//go:build integration

// This integration test runs the full migration set against a real libSQL
// server (Turso-compatible) in a container. Run with:
//
//	go test -tags=integration ./internal/platform/migrate/...
//
// It requires a working Docker daemon (testcontainers-go).
package migrate

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	idb "github.com/lumni/mirante/internal/platform/db"
)

func TestMigrationsAgainstLibSQL(t *testing.T) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "ghcr.io/tursodatabase/libsql-server:latest",
		ExposedPorts: []string{"8080/tcp"},
		Env:          map[string]string{"SQLD_NODE": "primary"},
		WaitingFor:   wait.ForListeningPort("8080/tcp").WithStartupTimeout(90 * time.Second),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	host, err := container.Host(ctx)
	require.NoError(t, err)
	port, err := container.MappedPort(ctx, "8080")
	require.NoError(t, err)

	url := fmt.Sprintf("http://%s:%s", host, port.Port())
	database, err := idb.Open(ctx, url, "")
	require.NoError(t, err)
	t.Cleanup(func() { _ = database.Close() })

	require.NoError(t, Up(database.DB))

	_, err = database.ExecContext(ctx,
		`INSERT INTO users (id, email, password_hash) VALUES (?, ?, ?)`,
		"u1", "owner@example.com", "argon2id-hash")
	require.NoError(t, err)

	var email string
	require.NoError(t, database.QueryRowContext(ctx,
		`SELECT email FROM users WHERE id = ?`, "u1").Scan(&email))
	require.Equal(t, "owner@example.com", email)
}
