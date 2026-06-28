// Package db opens a database/sql handle to SQLite (local file, pure-Go) or to
// libSQL/Turso (remote), chosen by the DATABASE_URL scheme. SQLite/libSQL is
// single-writer, so the pool is kept small and writes are serialized; callers
// use WithTx for atomic multi-statement work.
package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/tursodatabase/libsql-client-go/libsql" // driver: libsql
	_ "modernc.org/sqlite"                               // driver: sqlite (pure Go)
)

// DB embeds *sql.DB and adds transaction helpers.
type DB struct {
	*sql.DB
}

// Open resolves the driver from the URL scheme, opens the pool, applies pragmas
// and verifies connectivity.
func Open(ctx context.Context, url, authToken string) (*DB, error) {
	driver, dsn, err := resolve(url, authToken)
	if err != nil {
		return nil, err
	}

	sqlDB, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("open db (%s): %w", driver, err)
	}

	// Single-writer engine: one open connection avoids SQLITE_BUSY entirely.
	// Adequate for a single-user app; revisit if read concurrency ever matters.
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(pingCtx); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("ping db (%s): %w", driver, err)
	}

	return &DB{DB: sqlDB}, nil
}

// WithTx runs fn inside a transaction, committing on success and rolling back on
// error or panic.
func (db *DB) WithTx(ctx context.Context, fn func(*sql.Tx) error) (err error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	if err = fn(tx); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

// resolve maps a DATABASE_URL to a (driver, dsn) pair.
func resolve(url, authToken string) (driver, dsn string, err error) {
	switch {
	case url == ":memory:":
		return "sqlite", "file::memory:?cache=shared&" + filePragmas(false), nil
	case strings.HasPrefix(url, "file:"), strings.HasSuffix(url, ".db"):
		return "sqlite", appendQuery(url, filePragmas(true)), nil
	case strings.HasPrefix(url, "libsql://"),
		strings.HasPrefix(url, "http://"),
		strings.HasPrefix(url, "https://"),
		strings.HasPrefix(url, "ws://"),
		strings.HasPrefix(url, "wss://"):
		if authToken != "" {
			url = appendQuery(url, "authToken="+authToken)
		}
		return "libsql", url, nil
	default:
		return "", "", fmt.Errorf("unsupported DATABASE_URL scheme: %q", url)
	}
}

// filePragmas are modernc/sqlite connection pragmas. WAL is only valid for a
// real file (not :memory:).
func filePragmas(wal bool) string {
	p := "_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)"
	if wal {
		p += "&_pragma=journal_mode(WAL)"
	}
	return p
}

func appendQuery(url, q string) string {
	sep := "?"
	if strings.Contains(url, "?") {
		sep = "&"
	}
	return url + sep + q
}
