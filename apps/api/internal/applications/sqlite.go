package applications

import (
	"context"
	"database/sql"
	"errors"
	"time"

	idb "github.com/lumni/mirante/internal/platform/db"
)

type sqliteRepo struct{ db *idb.DB }

// NewSQLiteRepo builds a SQLite-backed application repository.
func NewSQLiteRepo(d *idb.DB) Repository { return &sqliteRepo{db: d} }

type rowScanner interface{ Scan(dest ...any) error }

const appCols = `id, job_id, titulo, empresa, status, notas, proxima_acao, data_acao, created_at, updated_at`

func scanApplication(s rowScanner) (*Application, error) {
	var (
		a                                             Application
		idStr, status                                 string
		jobID, titulo, empresa, notas, prox, dataAcao sql.NullString
		createdAt, updatedAt                          string
	)
	if err := s.Scan(&idStr, &jobID, &titulo, &empresa, &status, &notas, &prox, &dataAcao,
		&createdAt, &updatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	a.ID = ID(idStr)
	a.JobID = jobID.String
	a.Titulo = titulo.String
	a.Empresa = empresa.String
	a.Status = Status(status)
	a.Notas = notas.String
	a.ProximaAcao = prox.String
	a.DataAcao = dataAcao.String
	a.CreatedAt = idb.ParseTime(createdAt)
	a.UpdatedAt = idb.ParseTime(updatedAt)
	return &a, nil
}

func (r *sqliteRepo) Create(ctx context.Context, a *Application) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO applications (id, job_id, titulo, empresa, status, notas, proxima_acao, data_acao)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		string(a.ID), nullable(a.JobID), nullable(a.Titulo), nullable(a.Empresa), string(a.Status),
		nullable(a.Notas), nullable(a.ProximaAcao), nullable(a.DataAcao))
	return err
}

func (r *sqliteRepo) Get(ctx context.Context, id ID) (*Application, error) {
	return scanApplication(r.db.QueryRowContext(ctx,
		`SELECT `+appCols+` FROM applications WHERE id = ?`, string(id)))
}

func (r *sqliteRepo) List(ctx context.Context, f ListFilter) ([]*Application, error) {
	query := `SELECT ` + appCols + ` FROM applications`
	var args []any
	if f.Status != "" {
		query += ` WHERE status = ?`
		args = append(args, f.Status)
	}
	query += ` ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	out := []*Application{}
	for rows.Next() {
		a, err := scanApplication(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *sqliteRepo) Update(ctx context.Context, a *Application) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE applications SET titulo = ?, empresa = ?, status = ?, notas = ?,
		 proxima_acao = ?, data_acao = ?, updated_at = ? WHERE id = ?`,
		nullable(a.Titulo), nullable(a.Empresa), string(a.Status), nullable(a.Notas),
		nullable(a.ProximaAcao), nullable(a.DataAcao), idb.FormatTime(time.Now()), string(a.ID))
	if err != nil {
		return err
	}
	return mustAffect(res)
}

func (r *sqliteRepo) Delete(ctx context.Context, id ID) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM applications WHERE id = ?`, string(id))
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
