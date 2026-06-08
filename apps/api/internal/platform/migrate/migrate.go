// Package migrate runs the embedded goose migrations against an open database.
// The "sqlite3" dialect is used for both the pure-Go SQLite driver and the
// libSQL driver (libSQL is SQLite-compatible).
package migrate

import (
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"

	migfs "github.com/lumni/mirante/db"
)

const dir = "migrations"

func init() {
	goose.SetBaseFS(migfs.FS)
	if err := goose.SetDialect("sqlite3"); err != nil {
		panic(fmt.Sprintf("migrate: set dialect: %v", err))
	}
}

// Up applies all pending migrations.
func Up(db *sql.DB) error {
	if err := goose.Up(db, dir); err != nil {
		return fmt.Errorf("migrate up: %w", err)
	}
	return nil
}

// Down rolls back the most recent migration.
func Down(db *sql.DB) error {
	if err := goose.Down(db, dir); err != nil {
		return fmt.Errorf("migrate down: %w", err)
	}
	return nil
}

// Reset rolls every migration back to zero (used by integration tests).
func Reset(db *sql.DB) error {
	if err := goose.DownTo(db, dir, 0); err != nil {
		return fmt.Errorf("migrate reset: %w", err)
	}
	return nil
}
