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

// connMaxLifetime bounds how long a single remote (libSQL/Turso) connection is
// reused before it is recycled, so a stream the Turso edge has dropped
// server-side never lingers in the pool long enough to surface as a query error.
const connMaxLifetime = 5 * time.Minute

// connMaxIdleTime is kept under the libSQL/sqld hrana stream idle expiry (~10s) so
// a pooled connection never outlives its server-side stream and then fails on reuse
// ("stream is expired" — the dev sqld logs this every monitor reconcile otherwise).
// The reconnect cost is negligible at this app's traffic.
const connMaxIdleTime = 8 * time.Second

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

	if driver == "libsql" {
		// Remote Turso. The server serializes writes, so the local single-writer
		// rule (which exists only to avoid SQLITE_BUSY on a file) does not apply.
		// A small pool keeps one stale/poisoned connection from blocking every
		// query and lets a retry land on a healthy one. Capping the connection
		// lifetime recycles streams proactively, before the Turso edge recycles
		// them under us and the next query fails with "stream is closed".
		sqlDB.SetMaxOpenConns(4)
		sqlDB.SetConnMaxLifetime(connMaxLifetime)
	} else {
		// Local SQLite file: one writer avoids SQLITE_BUSY entirely.
		sqlDB.SetMaxOpenConns(1)
	}
	sqlDB.SetConnMaxIdleTime(connMaxIdleTime)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(pingCtx); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("ping db (%s): %w", driver, err)
	}

	return &DB{DB: sqlDB}, nil
}

// WithTx runs fn inside a transaction, committing on success and rolling back on
// error or panic. A transient connectivity failure before the commit is retried
// on a fresh connection (the rollback means nothing was applied, so re-running fn
// is safe). A failure during the commit itself is NOT retried: it is ambiguous —
// Turso may have committed before the stream dropped — and replaying it would
// double-apply non-idempotent writes (e.g. the rollup counters in Compact).
func (db *DB) WithTx(ctx context.Context, fn func(*sql.Tx) error) error {
	var err error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 && !backoff(ctx, attempt) {
			return ctx.Err()
		}
		var committed bool
		if committed, err = db.runTx(ctx, fn); err == nil {
			return nil
		}
		if committed || !IsTransient(err) {
			return err
		}
	}
	return err
}

// runTx runs one transaction attempt, reporting whether it reached the commit so
// the caller knows a returned error is past the safe-to-retry point.
func (db *DB) runTx(ctx context.Context, fn func(*sql.Tx) error) (committed bool, err error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()
	if err = fn(tx); err != nil {
		_ = tx.Rollback()
		return false, err
	}
	if err = tx.Commit(); err != nil {
		return true, fmt.Errorf("commit tx: %w", err)
	}
	return true, nil
}

// QueryContext shadows the embedded *sql.DB so every repo read gets connection
// resilience for free: a libSQL stream the Turso edge recycled server-side
// surfaces when the query is sent, and re-running a read on a fresh connection is
// always safe. Without this, a transient "stream is closed" on a pooled idle
// connection bubbles up as a 500 (the dashboard, which fires several reads at
// once, is the worst hit). Errors that arise during row iteration after a
// successful open are not retried, which is fine for the small result sets here.
func (db *DB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	var rows *sql.Rows
	err := Retry(ctx, func() error {
		var qerr error
		rows, qerr = db.DB.QueryContext(ctx, query, args...)
		return qerr
	})
	return rows, err
}

// QueryRowContext mirrors QueryContext for single-row reads. It re-runs the query
// on a transient failure and surfaces the chosen row via a closure so the retry
// can re-issue it; the caller scans the returned row as usual.
func (db *DB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	var row *sql.Row
	_ = Retry(ctx, func() error {
		row = db.DB.QueryRowContext(ctx, query, args...)
		return row.Err()
	})
	return row
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
