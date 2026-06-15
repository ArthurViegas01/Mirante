package jobs

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	idb "github.com/lumni/mirante/internal/platform/db"
	"github.com/lumni/mirante/internal/platform/tenant"
)

type sqliteRepo struct{ db *idb.DB }

// NewSQLiteRepo builds a SQLite-backed job repository.
func NewSQLiteRepo(d *idb.DB) Repository { return &sqliteRepo{db: d} }

type rowScanner interface{ Scan(dest ...any) error }

const jobCols = `id, titulo, empresa, descricao, url, localizacao, modelo, senioridade, resumo, created_at, updated_at`

func scanJob(s rowScanner) (*Job, error) {
	var (
		j                                                         Job
		idStr, titulo, modelo                                     string
		empresa, descricao, url, localizacao, senioridade, resumo sql.NullString
		createdAt, updatedAt                                      string
	)
	if err := s.Scan(&idStr, &titulo, &empresa, &descricao, &url, &localizacao, &modelo,
		&senioridade, &resumo, &createdAt, &updatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	j.ID = ID(idStr)
	j.Titulo = titulo
	j.Modelo = Modelo(modelo)
	j.Empresa = empresa.String
	j.Descricao = descricao.String
	j.URL = url.String
	j.Localizacao = localizacao.String
	j.Senioridade = senioridade.String
	j.Resumo = resumo.String
	j.CreatedAt = idb.ParseTime(createdAt)
	j.UpdatedAt = idb.ParseTime(updatedAt)
	j.Skills = []string{}
	return &j, nil
}

func (r *sqliteRepo) Create(ctx context.Context, j *Job) error {
	uid, _ := tenant.UserID(ctx)
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO jobs (id, user_id, titulo, empresa, descricao, url, localizacao, modelo, senioridade, resumo)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		string(j.ID), uid, j.Titulo, nullable(j.Empresa), nullable(j.Descricao), nullable(j.URL),
		nullable(j.Localizacao), string(j.Modelo), nullable(j.Senioridade), nullable(j.Resumo))
	return err
}

func (r *sqliteRepo) Get(ctx context.Context, id ID) (*Job, error) {
	uid, _ := tenant.UserID(ctx)
	j, err := scanJob(r.db.QueryRowContext(ctx,
		`SELECT `+jobCols+` FROM jobs WHERE id = ? AND user_id = ?`, string(id), uid))
	if err != nil {
		return nil, err
	}
	if j.Skills, err = r.listSkills(ctx, id); err != nil {
		return nil, err
	}
	return j, nil
}

func (r *sqliteRepo) List(ctx context.Context) ([]*Job, error) {
	uid, _ := tenant.UserID(ctx)
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+jobCols+` FROM jobs WHERE user_id = ? ORDER BY created_at DESC`, uid)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	out := []*Job{}
	for rows.Next() {
		j, err := scanJob(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, j)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for _, j := range out {
		if j.Skills, err = r.listSkills(ctx, j.ID); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func (r *sqliteRepo) Update(ctx context.Context, j *Job) error {
	uid, _ := tenant.UserID(ctx)
	res, err := r.db.ExecContext(ctx,
		`UPDATE jobs SET titulo = ?, empresa = ?, descricao = ?, url = ?, localizacao = ?,
		 modelo = ?, senioridade = ?, resumo = ?, updated_at = ? WHERE id = ? AND user_id = ?`,
		j.Titulo, nullable(j.Empresa), nullable(j.Descricao), nullable(j.URL), nullable(j.Localizacao),
		string(j.Modelo), nullable(j.Senioridade), nullable(j.Resumo), idb.FormatTime(time.Now()), string(j.ID), uid)
	if err != nil {
		return err
	}
	return mustAffect(res)
}

func (r *sqliteRepo) Delete(ctx context.Context, id ID) error {
	uid, _ := tenant.UserID(ctx)
	res, err := r.db.ExecContext(ctx, `DELETE FROM jobs WHERE id = ? AND user_id = ?`, string(id), uid)
	if err != nil {
		return err
	}
	return mustAffect(res)
}

func (r *sqliteRepo) SetSkills(ctx context.Context, jobID ID, skillNames []string) error {
	uid, _ := tenant.UserID(ctx)
	return r.db.WithTx(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx,
			`DELETE FROM job_skills WHERE job_id = ? AND user_id = ?`, string(jobID), uid); err != nil {
			return err
		}
		for _, name := range skillNames {
			name = strings.TrimSpace(name)
			if name == "" {
				continue
			}
			if _, err := tx.ExecContext(ctx,
				`INSERT OR IGNORE INTO job_skills (job_id, skill, user_id) VALUES (?, ?, ?)`,
				string(jobID), name, uid); err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *sqliteRepo) listSkills(ctx context.Context, jobID ID) ([]string, error) {
	uid, _ := tenant.UserID(ctx)
	rows, err := r.db.QueryContext(ctx,
		`SELECT skill FROM job_skills WHERE job_id = ? AND user_id = ? ORDER BY skill`, string(jobID), uid)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	out := []string{}
	for rows.Next() {
		var skill string
		if err := rows.Scan(&skill); err != nil {
			return nil, err
		}
		out = append(out, skill)
	}
	return out, rows.Err()
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
