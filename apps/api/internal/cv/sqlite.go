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
	if errors.Is(err, sql.ErrNoRows) {
		return &Profile{}, nil
	}
	if err != nil {
		return nil, err
	}
	p.Nome = nome.String
	p.Titulo = titulo.String
	p.TituloAlvo = tituloAlvo.String
	p.Resumo = resumo.String
	if updatedAt.Valid {
		p.UpdatedAt = idb.ParseTime(updatedAt.String)
	}
	return &p, nil
}

func (r *sqliteRepo) SaveProfile(ctx context.Context, p *Profile) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO cv_profile (id, nome, titulo, titulo_alvo, resumo, updated_at)
		 VALUES (?, ?, ?, ?, ?, strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
		 ON CONFLICT(id) DO UPDATE SET
		   nome = excluded.nome, titulo = excluded.titulo, titulo_alvo = excluded.titulo_alvo,
		   resumo = excluded.resumo, updated_at = excluded.updated_at`,
		profileID, nullable(p.Nome), nullable(p.Titulo), nullable(p.TituloAlvo), nullable(p.Resumo))
	return err
}

func nullable(s string) any {
	if s == "" {
		return nil
	}
	return s
}
