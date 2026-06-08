package projects

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	idb "github.com/lumni/mirante/internal/platform/db"
	idgen "github.com/lumni/mirante/internal/platform/id"
)

type sqliteRepo struct{ db *idb.DB }

// NewSQLiteRepo builds a SQLite-backed project repository.
func NewSQLiteRepo(d *idb.DB) Repository { return &sqliteRepo{db: d} }

type rowScanner interface{ Scan(dest ...any) error }

const projectCols = `id, nome, codinome, descricao, repo, status, visibilidade, created_at, updated_at`

func scanProject(s rowScanner) (*Project, error) {
	var (
		p                         Project
		idStr, status, vis        string
		codinome, descricao, repo sql.NullString
		createdAt, updatedAt      string
	)
	if err := s.Scan(&idStr, &p.Nome, &codinome, &descricao, &repo, &status, &vis, &createdAt, &updatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	p.ID = ID(idStr)
	p.Status = Status(status)
	p.Visibilidade = Visibility(vis)
	p.Codinome = codinome.String
	p.Descricao = descricao.String
	p.Repo = repo.String
	p.CreatedAt = idb.ParseTime(createdAt)
	p.UpdatedAt = idb.ParseTime(updatedAt)
	p.Links = []Link{}
	p.Tags = []string{}
	return &p, nil
}

func (r *sqliteRepo) Create(ctx context.Context, p *Project) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO projects (id, nome, codinome, descricao, repo, status, visibilidade)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		string(p.ID), p.Nome, nullable(p.Codinome), nullable(p.Descricao), nullable(p.Repo),
		string(p.Status), string(p.Visibilidade))
	return err
}

func (r *sqliteRepo) Get(ctx context.Context, id ID) (*Project, error) {
	p, err := scanProject(r.db.QueryRowContext(ctx,
		`SELECT `+projectCols+` FROM projects WHERE id = ?`, string(id)))
	if err != nil {
		return nil, err
	}
	if p.Links, err = r.listLinks(ctx, id); err != nil {
		return nil, err
	}
	if p.Tags, err = r.listTags(ctx, id); err != nil {
		return nil, err
	}
	return p, nil
}

func (r *sqliteRepo) List(ctx context.Context, f ListFilter) ([]*Project, error) {
	query := `SELECT ` + projectCols + ` FROM projects`
	var args []any
	if f.Status != "" {
		query += ` WHERE status = ?`
		args = append(args, f.Status)
	}
	query += ` ORDER BY updated_at DESC`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	out := []*Project{}
	for rows.Next() {
		p, err := scanProject(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	// Load tags per project (few projects; single user).
	for _, p := range out {
		if p.Tags, err = r.listTags(ctx, p.ID); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func (r *sqliteRepo) Update(ctx context.Context, p *Project) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE projects SET nome = ?, codinome = ?, descricao = ?, repo = ?, status = ?,
		 visibilidade = ?, updated_at = ? WHERE id = ?`,
		p.Nome, nullable(p.Codinome), nullable(p.Descricao), nullable(p.Repo),
		string(p.Status), string(p.Visibilidade), idb.FormatTime(time.Now()), string(p.ID))
	if err != nil {
		return err
	}
	return mustAffect(res)
}

func (r *sqliteRepo) Delete(ctx context.Context, id ID) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM projects WHERE id = ?`, string(id))
	if err != nil {
		return err
	}
	return mustAffect(res)
}

func (r *sqliteRepo) AddLink(ctx context.Context, l *Link) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO project_links (id, project_id, label, url, kind, sort_order)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		string(l.ID), string(l.ProjectID), l.Label, l.URL, l.Kind, l.SortOrder)
	return err
}

func (r *sqliteRepo) RemoveLink(ctx context.Context, projectID, linkID ID) error {
	res, err := r.db.ExecContext(ctx,
		`DELETE FROM project_links WHERE id = ? AND project_id = ?`, string(linkID), string(projectID))
	if err != nil {
		return err
	}
	return mustAffect(res)
}

func (r *sqliteRepo) SetTags(ctx context.Context, projectID ID, names []string) error {
	return r.db.WithTx(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `DELETE FROM project_tags WHERE project_id = ?`, string(projectID)); err != nil {
			return err
		}
		for _, name := range names {
			name = strings.TrimSpace(name)
			if name == "" {
				continue
			}
			if _, err := tx.ExecContext(ctx,
				`INSERT OR IGNORE INTO tags (id, name) VALUES (?, ?)`, idgen.New(), name); err != nil {
				return err
			}
			var tagID string
			if err := tx.QueryRowContext(ctx, `SELECT id FROM tags WHERE name = ?`, name).Scan(&tagID); err != nil {
				return err
			}
			if _, err := tx.ExecContext(ctx,
				`INSERT OR IGNORE INTO project_tags (project_id, tag_id) VALUES (?, ?)`,
				string(projectID), tagID); err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *sqliteRepo) listLinks(ctx context.Context, projectID ID) ([]Link, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, project_id, label, url, kind, sort_order, created_at
		 FROM project_links WHERE project_id = ? ORDER BY sort_order, created_at`, string(projectID))
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	links := []Link{}
	for rows.Next() {
		var (
			l                       Link
			idStr, pidStr, kind, ts string
		)
		if err := rows.Scan(&idStr, &pidStr, &l.Label, &l.URL, &kind, &l.SortOrder, &ts); err != nil {
			return nil, err
		}
		l.ID = ID(idStr)
		l.ProjectID = ID(pidStr)
		l.Kind = kind
		l.CreatedAt = idb.ParseTime(ts)
		links = append(links, l)
	}
	return links, rows.Err()
}

func (r *sqliteRepo) listTags(ctx context.Context, projectID ID) ([]string, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT t.name FROM tags t
		 JOIN project_tags pt ON pt.tag_id = t.id
		 WHERE pt.project_id = ? ORDER BY t.name`, string(projectID))
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	tags := []string{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		tags = append(tags, name)
	}
	return tags, rows.Err()
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
