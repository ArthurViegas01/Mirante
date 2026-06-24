package db

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"testing"
)

func TestIsTransient(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"stream closed", errors.New("failed to execute SQL:\nstream is closed: driver: bad connection"), true},
		{"edge 502", errors.New("failed to execute SQL:\nerror code 502: connect to upstream failed"), true},
		{"wrapped bad conn", fmt.Errorf("begin tx: %w", driver.ErrBadConn), true},
		{"eof", io.EOF, true},
		{"conn reset", errors.New("read tcp 10.0.0.1:443: connection reset by peer"), true},
		{"i/o timeout", errors.New("dial tcp: i/o timeout"), true},
		{"context canceled", context.Canceled, false},
		{"deadline exceeded", context.DeadlineExceeded, false},
		{"no rows", sql.ErrNoRows, false},
		{"syntax error", errors.New(`near "slect": syntax error`), false},
		{"unique violation", errors.New("UNIQUE constraint failed: users.email"), false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := IsTransient(c.err); got != c.want {
				t.Fatalf("IsTransient(%v) = %v, want %v", c.err, got, c.want)
			}
		})
	}
}

func TestRetrySucceedsAfterTransient(t *testing.T) {
	calls := 0
	err := Retry(context.Background(), func() error {
		calls++
		if calls < 3 {
			return driver.ErrBadConn
		}
		return nil
	})
	if err != nil {
		t.Fatalf("want nil error, got %v", err)
	}
	if calls != 3 {
		t.Fatalf("want 3 calls, got %d", calls)
	}
}

func TestRetryReturnsPermanentImmediately(t *testing.T) {
	sentinel := errors.New("boom")
	calls := 0
	err := Retry(context.Background(), func() error {
		calls++
		return sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("want sentinel, got %v", err)
	}
	if calls != 1 {
		t.Fatalf("permanent error must not retry: got %d calls", calls)
	}
}

func TestRetryExhaustsAttempts(t *testing.T) {
	calls := 0
	err := Retry(context.Background(), func() error {
		calls++
		return driver.ErrBadConn
	})
	if !errors.Is(err, driver.ErrBadConn) {
		t.Fatalf("want last transient error, got %v", err)
	}
	if calls != maxAttempts {
		t.Fatalf("want %d calls, got %d", maxAttempts, calls)
	}
}

func TestRetryStopsWhenContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	calls := 0
	err := Retry(ctx, func() error {
		calls++
		return driver.ErrBadConn
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("want context.Canceled, got %v", err)
	}
	// The first attempt runs; the backoff before the second sees the dead context.
	if calls != 1 {
		t.Fatalf("want 1 call before bailing, got %d", calls)
	}
}

func TestWithTxRetriesPreCommitTransient(t *testing.T) {
	d := openMemory(t)
	ctx := context.Background()
	if _, err := d.ExecContext(ctx, `CREATE TABLE t (id INTEGER PRIMARY KEY)`); err != nil {
		t.Fatal(err)
	}

	calls := 0
	err := d.WithTx(ctx, func(tx *sql.Tx) error {
		calls++
		if calls < 2 {
			return driver.ErrBadConn // pre-commit transient on the first attempt
		}
		_, e := tx.ExecContext(ctx, `INSERT INTO t (id) VALUES (1)`)
		return e
	})
	if err != nil {
		t.Fatalf("want nil error, got %v", err)
	}
	if calls != 2 {
		t.Fatalf("want 2 attempts, got %d", calls)
	}
	if got := countRows(t, d); got != 1 {
		t.Fatalf("want exactly 1 row (first attempt rolled back), got %d", got)
	}
}

func TestWithTxDoesNotRetryPermanent(t *testing.T) {
	d := openMemory(t)
	ctx := context.Background()
	if _, err := d.ExecContext(ctx, `CREATE TABLE t (id INTEGER PRIMARY KEY)`); err != nil {
		t.Fatal(err)
	}

	sentinel := errors.New("nope")
	calls := 0
	err := d.WithTx(ctx, func(tx *sql.Tx) error {
		calls++
		if _, e := tx.ExecContext(ctx, `INSERT INTO t (id) VALUES (2)`); e != nil {
			return e
		}
		return sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("want sentinel, got %v", err)
	}
	if calls != 1 {
		t.Fatalf("permanent error must not retry: got %d attempts", calls)
	}
	if got := countRows(t, d); got != 0 {
		t.Fatalf("want 0 rows (rolled back), got %d", got)
	}
}

func openMemory(t *testing.T) *DB {
	t.Helper()
	d, err := Open(context.Background(), ":memory:", "")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = d.Close() })
	return d
}

func countRows(t *testing.T, d *DB) int {
	t.Helper()
	var n int
	if err := d.QueryRowContext(context.Background(), `SELECT COUNT(*) FROM t`).Scan(&n); err != nil {
		t.Fatalf("count: %v", err)
	}
	return n
}
