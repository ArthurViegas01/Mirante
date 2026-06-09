package cv

import (
	"context"
	"database/sql"
	"errors"

	idb "github.com/lumni/mirante/internal/platform/db"
)

// profileID is the fixed key of the singleton row.
const profileID = "owner"

type sqliteRepo struct{ db *idb.DB }

// NewSQLiteRepo builds a SQLite-backed CV repository.
func NewSQLiteRepo(d *idb.DB) Repository { return &sqliteRepo{db: d} }

func (r *sqliteRepo) GetProfile(ctx context.Context) (*Profile, error) {
	var (
		nome, titulo, tituloAlvo, resumo, updatedAt sql.NullString
		p                                           Profile
	)
	err := r.db.QueryRowContext(ctx,
		`SELECT nome, titulo, titulo_alvo, resumo, updated_at FROM cv_profile WHERE id = ?`, profileID).
		Scan(&nome, &titulo, &tituloAlvo, &resumo, &updatedAt)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		// No profile row yet — leave the identity blank but still load any skills.
	case err != nil:
		return nil, err
	default:
		p.Nome = nome.String
		p.Titulo = titulo.String
		p.TituloAlvo = tituloAlvo.String
		p.Resumo = resumo.String
		if updatedAt.Valid {
			p.UpdatedAt = idb.ParseTime(updatedAt.String)
		}
	}
	if p.Skills, err = r.listSkills(ctx); err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *sqliteRepo) listSkills(ctx context.Context) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT skill FROM cv_skills ORDER BY skill`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := []string{}
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *sqliteRepo) SaveProfile(ctx context.Context, p *Profile) error {
	return r.db.WithTx(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO cv_profile (id, nome, titulo, titulo_alvo, resumo, updated_at)
			 VALUES (?, ?, ?, ?, ?, strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
			 ON CONFLICT(id) DO UPDATE SET
			   nome = excluded.nome, titulo = excluded.titulo, titulo_alvo = excluded.titulo_alvo,
			   resumo = excluded.resumo, updated_at = excluded.updated_at`,
			profileID, nullable(p.Nome), nullable(p.Titulo), nullable(p.TituloAlvo), nullable(p.Resumo)); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM cv_skills`); err != nil {
			return err
		}
		for _, sk := range p.Skills {
			if _, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO cv_skills (skill) VALUES (?)`, sk); err != nil {
				return err
			}
		}
		return nil
	})
}

func nullable(s string) any {
	if s == "" {
		return nil
	}
	return s
}
