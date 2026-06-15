package cv

import (
	"context"
	"database/sql"
	"errors"

	idb "github.com/lumni/mirante/internal/platform/db"
	idgen "github.com/lumni/mirante/internal/platform/id"
	"github.com/lumni/mirante/internal/platform/tenant"
)

type sqliteRepo struct{ db *idb.DB }

// NewSQLiteRepo builds a SQLite-backed CV repository.
func NewSQLiteRepo(d *idb.DB) Repository { return &sqliteRepo{db: d} }

func (r *sqliteRepo) GetProfile(ctx context.Context) (*Profile, error) {
	uid, _ := tenant.UserID(ctx)
	var (
		nome, titulo, tituloAlvo, contato, resumo, updatedAt sql.NullString
		p                                                    Profile
	)
	err := r.db.QueryRowContext(ctx,
		`SELECT nome, titulo, titulo_alvo, contato, resumo, updated_at FROM cv_profile WHERE user_id = ?`, uid).
		Scan(&nome, &titulo, &tituloAlvo, &contato, &resumo, &updatedAt)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		// No profile row yet — leave the identity blank but still load sub-lists.
	case err != nil:
		return nil, err
	default:
		p.Nome = nome.String
		p.Titulo = titulo.String
		p.TituloAlvo = tituloAlvo.String
		p.Contato = contato.String
		p.Resumo = resumo.String
		if updatedAt.Valid {
			p.UpdatedAt = idb.ParseTime(updatedAt.String)
		}
	}
	if p.Skills, err = r.listSkills(ctx); err != nil {
		return nil, err
	}
	if p.Experiences, err = r.listExperiences(ctx); err != nil {
		return nil, err
	}
	if p.Education, err = r.listEducation(ctx); err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *sqliteRepo) listSkills(ctx context.Context) ([]string, error) {
	uid, _ := tenant.UserID(ctx)
	rows, err := r.db.QueryContext(ctx, `SELECT skill FROM cv_skills WHERE user_id = ? ORDER BY skill`, uid)
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

func (r *sqliteRepo) listExperiences(ctx context.Context) ([]Experience, error) {
	uid, _ := tenant.UserID(ctx)
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, empresa, cargo, inicio, fim, descricao FROM cv_experience WHERE user_id = ? ORDER BY ordem, id`, uid)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := []Experience{}
	for rows.Next() {
		var (
			e                                      Experience
			empresa, cargo, inicio, fim, descricao sql.NullString
		)
		if err := rows.Scan(&e.ID, &empresa, &cargo, &inicio, &fim, &descricao); err != nil {
			return nil, err
		}
		e.Empresa, e.Cargo, e.Inicio, e.Fim, e.Descricao =
			empresa.String, cargo.String, inicio.String, fim.String, descricao.String
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *sqliteRepo) listEducation(ctx context.Context) ([]Education, error) {
	uid, _ := tenant.UserID(ctx)
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, instituicao, curso, inicio, fim FROM cv_education WHERE user_id = ? ORDER BY ordem, id`, uid)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := []Education{}
	for rows.Next() {
		var (
			e                               Education
			instituicao, curso, inicio, fim sql.NullString
		)
		if err := rows.Scan(&e.ID, &instituicao, &curso, &inicio, &fim); err != nil {
			return nil, err
		}
		e.Instituicao, e.Curso, e.Inicio, e.Fim =
			instituicao.String, curso.String, inicio.String, fim.String
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *sqliteRepo) SaveCV(ctx context.Context, p *Profile) error {
	uid, _ := tenant.UserID(ctx)
	return r.db.WithTx(ctx, func(tx *sql.Tx) error {
		// One profile row per user (ux_cv_profile_user); upsert keyed by user_id.
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO cv_profile (id, user_id, nome, titulo, titulo_alvo, contato, resumo, updated_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?, strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
			 ON CONFLICT(user_id) DO UPDATE SET
			   nome = excluded.nome, titulo = excluded.titulo, titulo_alvo = excluded.titulo_alvo,
			   contato = excluded.contato, resumo = excluded.resumo, updated_at = excluded.updated_at`,
			idgen.New(), uid, nullable(p.Nome), nullable(p.Titulo), nullable(p.TituloAlvo),
			nullable(p.Contato), nullable(p.Resumo)); err != nil {
			return err
		}

		if _, err := tx.ExecContext(ctx, `DELETE FROM cv_skills WHERE user_id = ?`, uid); err != nil {
			return err
		}
		for _, sk := range p.Skills {
			if _, err := tx.ExecContext(ctx,
				`INSERT OR IGNORE INTO cv_skills (user_id, skill) VALUES (?, ?)`, uid, sk); err != nil {
				return err
			}
		}

		if _, err := tx.ExecContext(ctx, `DELETE FROM cv_experience WHERE user_id = ?`, uid); err != nil {
			return err
		}
		for i, e := range p.Experiences {
			if _, err := tx.ExecContext(ctx,
				`INSERT INTO cv_experience (id, user_id, empresa, cargo, inicio, fim, descricao, ordem)
				 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
				idgen.New(), uid, nullable(e.Empresa), nullable(e.Cargo), nullable(e.Inicio),
				nullable(e.Fim), nullable(e.Descricao), i); err != nil {
				return err
			}
		}

		if _, err := tx.ExecContext(ctx, `DELETE FROM cv_education WHERE user_id = ?`, uid); err != nil {
			return err
		}
		for i, e := range p.Education {
			if _, err := tx.ExecContext(ctx,
				`INSERT INTO cv_education (id, user_id, instituicao, curso, inicio, fim, ordem)
				 VALUES (?, ?, ?, ?, ?, ?, ?)`,
				idgen.New(), uid, nullable(e.Instituicao), nullable(e.Curso), nullable(e.Inicio),
				nullable(e.Fim), i); err != nil {
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
