package subscriptions

import (
	"context"
	"database/sql"
	"errors"
	"time"

	idb "github.com/lumni/mirante/internal/platform/db"
	"github.com/lumni/mirante/internal/platform/tenant"
)

type sqliteRepo struct{ db *idb.DB }

// NewSQLiteRepo builds a SQLite-backed subscription repository.
func NewSQLiteRepo(d *idb.DB) Repository { return &sqliteRepo{db: d} }

type rowScanner interface{ Scan(dest ...any) error }

const subCols = `id, project_id, service_id, nome, provider, valor_cents, moeda, ciclo, ativo, notas, created_at, updated_at`

func scanSubscription(s rowScanner) (*Subscription, error) {
	var (
		sub                        Subscription
		idStr, projectID           string
		serviceID, provider, notas sql.NullString
		moeda, ciclo               string
		ativo                      int
		createdAt, updatedAt       string
	)
	if err := s.Scan(&idStr, &projectID, &serviceID, &sub.Nome, &provider, &sub.ValorCents,
		&moeda, &ciclo, &ativo, &notas, &createdAt, &updatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	sub.ID = ID(idStr)
	sub.ProjectID = projectID
	sub.ServiceID = serviceID.String
	sub.Provider = provider.String
	sub.Moeda = Currency(moeda)
	sub.Ciclo = Cycle(ciclo)
	sub.Ativo = ativo == 1
	sub.Notas = notas.String
	sub.CreatedAt = idb.ParseTime(createdAt)
	sub.UpdatedAt = idb.ParseTime(updatedAt)
	return &sub, nil
}

func (r *sqliteRepo) Create(ctx context.Context, s *Subscription) error {
	uid, _ := tenant.UserID(ctx)
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO subscriptions (id, user_id, project_id, service_id, nome, provider, valor_cents, moeda, ciclo, ativo, notas)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		string(s.ID), uid, s.ProjectID, nullable(s.ServiceID), s.Nome, nullable(s.Provider),
		s.ValorCents, string(s.Moeda), string(s.Ciclo), boolToInt(s.Ativo), nullable(s.Notas))
	return err
}

func (r *sqliteRepo) Get(ctx context.Context, id ID) (*Subscription, error) {
	uid, _ := tenant.UserID(ctx)
	return scanSubscription(r.db.QueryRowContext(ctx,
		`SELECT `+subCols+` FROM subscriptions WHERE id = ? AND user_id = ?`, string(id), uid))
}

func (r *sqliteRepo) List(ctx context.Context, f ListFilter) ([]*Subscription, error) {
	uid, _ := tenant.UserID(ctx)
	query := `SELECT ` + subCols + ` FROM subscriptions WHERE user_id = ?`
	args := []any{uid}
	if f.ProjectID != "" {
		query += ` AND project_id = ?`
		args = append(args, f.ProjectID)
	}
	query += ` ORDER BY created_at`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	out := []*Subscription{}
	for rows.Next() {
		sub, err := scanSubscription(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, sub)
	}
	return out, rows.Err()
}

func (r *sqliteRepo) Update(ctx context.Context, s *Subscription) error {
	uid, _ := tenant.UserID(ctx)
	res, err := r.db.ExecContext(ctx,
		`UPDATE subscriptions SET service_id = ?, nome = ?, provider = ?, valor_cents = ?,
		 moeda = ?, ciclo = ?, ativo = ?, notas = ?, updated_at = ? WHERE id = ? AND user_id = ?`,
		nullable(s.ServiceID), s.Nome, nullable(s.Provider), s.ValorCents,
		string(s.Moeda), string(s.Ciclo), boolToInt(s.Ativo), nullable(s.Notas),
		idb.FormatTime(time.Now()), string(s.ID), uid)
	if err != nil {
		return err
	}
	return mustAffect(res)
}

func (r *sqliteRepo) Delete(ctx context.Context, id ID) error {
	uid, _ := tenant.UserID(ctx)
	res, err := r.db.ExecContext(ctx, `DELETE FROM subscriptions WHERE id = ? AND user_id = ?`, string(id), uid)
	if err != nil {
		return err
	}
	return mustAffect(res)
}

func mustAffect(res sql.Result) error {
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func nullable(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
