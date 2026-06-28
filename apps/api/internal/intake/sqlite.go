package intake

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

// NewSQLiteRepo builds a SQLite-backed intake repository.
func NewSQLiteRepo(d *idb.DB) Repository { return &sqliteRepo{db: d} }

type rowScanner interface{ Scan(dest ...any) error }

const itemCols = `id, fonte, fonte_id, titulo, categoria, nivel, publicado, tempo_restante,
	restante_horas, propostas, interessados, teaser, url, enviar_url, skills, score, estado,
	created_at, updated_at`

func scanItem(s rowScanner) (*Item, error) {
	var (
		it                                                                   Item
		idStr, fonte, fonteID, titulo, estado                                string
		categoria, nivel, publicado, tempoRestante, teaser, url, enviar, sks sql.NullString
		createdAt, updatedAt                                                 string
	)
	if err := s.Scan(&idStr, &fonte, &fonteID, &titulo, &categoria, &nivel, &publicado, &tempoRestante,
		&it.RestanteHoras, &it.Propostas, &it.Interessados, &teaser, &url, &enviar, &sks, &it.Score, &estado,
		&createdAt, &updatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	it.ID = ID(idStr)
	it.Fonte = fonte
	it.FonteID = fonteID
	it.Titulo = titulo
	it.Estado = Estado(estado)
	it.Categoria = categoria.String
	it.Nivel = nivel.String
	it.Publicado = publicado.String
	it.TempoRestante = tempoRestante.String
	it.Teaser = teaser.String
	it.URL = url.String
	it.EnviarURL = enviar.String
	it.Skills = splitCSV(sks.String)
	it.CreatedAt = idb.ParseTime(createdAt)
	it.UpdatedAt = idb.ParseTime(updatedAt)
	return &it, nil
}

func (r *sqliteRepo) Upsert(ctx context.Context, it *Item) (bool, error) {
	uid, _ := tenant.UserID(ctx)
	res, err := r.db.ExecContext(ctx,
		`INSERT OR IGNORE INTO intake_items
			(id, user_id, fonte, fonte_id, titulo, categoria, nivel, publicado, tempo_restante,
			 restante_horas, propostas, interessados, teaser, url, enviar_url, skills, score, estado)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		string(it.ID), uid, it.Fonte, it.FonteID, it.Titulo, nullable(it.Categoria), nullable(it.Nivel),
		nullable(it.Publicado), nullable(it.TempoRestante), it.RestanteHoras, it.Propostas, it.Interessados,
		nullable(it.Teaser), nullable(it.URL), nullable(it.EnviarURL), nullable(strings.Join(it.Skills, ",")),
		it.Score, string(it.Estado))
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}

func (r *sqliteRepo) List(ctx context.Context, f ListFilter) ([]*Item, error) {
	uid, _ := tenant.UserID(ctx)
	q := `SELECT ` + itemCols + ` FROM intake_items WHERE user_id = ?`
	args := []any{uid}
	if f.Estado != "" {
		q += ` AND estado = ?`
		args = append(args, string(f.Estado))
	}
	if f.MinScore > 0 {
		q += ` AND score >= ?`
		args = append(args, f.MinScore)
	}
	q += ` ORDER BY score DESC, created_at DESC`

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	out := []*Item{}
	for rows.Next() {
		it, err := scanItem(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

func (r *sqliteRepo) Get(ctx context.Context, id ID) (*Item, error) {
	uid, _ := tenant.UserID(ctx)
	return scanItem(r.db.QueryRowContext(ctx,
		`SELECT `+itemCols+` FROM intake_items WHERE id = ? AND user_id = ?`, string(id), uid))
}

func (r *sqliteRepo) SetEstado(ctx context.Context, id ID, estado Estado) error {
	uid, _ := tenant.UserID(ctx)
	res, err := r.db.ExecContext(ctx,
		`UPDATE intake_items SET estado = ?, updated_at = ? WHERE id = ? AND user_id = ?`,
		string(estado), idb.FormatTime(time.Now()), string(id), uid)
	if err != nil {
		return err
	}
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

func splitCSV(s string) []string {
	out := []string{}
	for _, p := range strings.Split(s, ",") {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}
