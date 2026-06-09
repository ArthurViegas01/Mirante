package llm

import (
	"context"

	idb "github.com/lumni/mirante/internal/platform/db"
)

type sqliteLedger struct{ db *idb.DB }

// NewSQLiteLedger persists usage rows to the llm_usage table.
func NewSQLiteLedger(d *idb.DB) Ledger { return &sqliteLedger{db: d} }

func (l *sqliteLedger) Record(ctx context.Context, e UsageEntry) error {
	_, err := l.db.ExecContext(ctx,
		`INSERT INTO llm_usage (provider, model, route, input_tokens, output_tokens)
		 VALUES (?, ?, ?, ?, ?)`,
		e.Provider, e.Model, e.Route, e.InputTokens, e.OutputTokens)
	return err
}
