package tasks

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	idb "github.com/lumni/mirante/internal/platform/db"
	idgen "github.com/lumni/mirante/internal/platform/id"
	"github.com/lumni/mirante/internal/platform/tenant"
)

type sqliteRepo struct{ db *idb.DB }

// NewSQLiteRepo builds a SQLite-backed task repository.
func NewSQLiteRepo(d *idb.DB) Repository { return &sqliteRepo{db: d} }

type rowScanner interface{ Scan(dest ...any) error }

const taskCols = `id, titulo, descricao, status, prioridade, prazo, project_id, job_id, created_at, updated_at`

func scanTask(s rowScanner) (*Task, error) {
	var (
		t                         Task
		idStr, status, prioridade string
		descricao, prazo          sql.NullString
		projectID, jobID          sql.NullString
		createdAt, updatedAt      string
	)
	if err := s.Scan(&idStr, &t.Titulo, &descricao, &status, &prioridade, &prazo,
		&projectID, &jobID, &createdAt, &updatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	t.ID = ID(idStr)
	t.Status = Status(status)
	t.Prioridade = Priority(prioridade)
	t.Descricao = descricao.String
	t.Prazo = prazo.String
	t.ProjectID = projectID.String
	t.JobID = jobID.String
	t.CreatedAt = idb.ParseTime(createdAt)
	t.UpdatedAt = idb.ParseTime(updatedAt)
	t.Tags = []string{}
	return &t, nil
}

func (r *sqliteRepo) Create(ctx context.Context, t *Task) error {
	uid, _ := tenant.UserID(ctx)
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO tasks (id, user_id, titulo, descricao, status, prioridade, prazo, project_id, job_id)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		string(t.ID), uid, t.Titulo, nullable(t.Descricao), string(t.Status), string(t.Prioridade),
		nullable(t.Prazo), nullable(t.ProjectID), nullable(t.JobID))
	return err
}

func (r *sqliteRepo) Get(ctx context.Context, id ID) (*Task, error) {
	uid, _ := tenant.UserID(ctx)
	t, err := scanTask(r.db.QueryRowContext(ctx,
		`SELECT `+taskCols+` FROM tasks WHERE id = ? AND user_id = ?`, string(id), uid))
	if err != nil {
		return nil, err
	}
	if t.Tags, err = r.listTags(ctx, id); err != nil {
		return nil, err
	}
	return t, nil
}

func (r *sqliteRepo) List(ctx context.Context, f ListFilter) ([]*Task, error) {
	uid, _ := tenant.UserID(ctx)
	query := `SELECT ` + taskCols + ` FROM tasks`
	conds := []string{"user_id = ?"}
	args := []any{uid}
	if f.Status != "" {
		conds = append(conds, "status = ?")
		args = append(args, f.Status)
	}
	if f.ProjectID != "" {
		conds = append(conds, "project_id = ?")
		args = append(args, f.ProjectID)
	}
	query += ` WHERE ` + strings.Join(conds, " AND ")
	query += ` ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	out := []*Task{}
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	// Load tags per task (single user; few rows).
	for _, t := range out {
		if t.Tags, err = r.listTags(ctx, t.ID); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func (r *sqliteRepo) Update(ctx context.Context, t *Task) error {
	uid, _ := tenant.UserID(ctx)
	res, err := r.db.ExecContext(ctx,
		`UPDATE tasks SET titulo = ?, descricao = ?, status = ?, prioridade = ?, prazo = ?,
		 project_id = ?, job_id = ?, updated_at = ? WHERE id = ? AND user_id = ?`,
		t.Titulo, nullable(t.Descricao), string(t.Status), string(t.Prioridade), nullable(t.Prazo),
		nullable(t.ProjectID), nullable(t.JobID), idb.FormatTime(time.Now()), string(t.ID), uid)
	if err != nil {
		return err
	}
	return mustAffect(res)
}

func (r *sqliteRepo) Delete(ctx context.Context, id ID) error {
	uid, _ := tenant.UserID(ctx)
	res, err := r.db.ExecContext(ctx, `DELETE FROM tasks WHERE id = ? AND user_id = ?`, string(id), uid)
	if err != nil {
		return err
	}
	return mustAffect(res)
}

func (r *sqliteRepo) SetTags(ctx context.Context, taskID ID, names []string) error {
	uid, _ := tenant.UserID(ctx)
	return r.db.WithTx(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx,
			`DELETE FROM task_tags WHERE task_id = ? AND user_id = ?`, string(taskID), uid); err != nil {
			return err
		}
		for _, name := range names {
			name = strings.TrimSpace(name)
			if name == "" {
				continue
			}
			if _, err := tx.ExecContext(ctx,
				`INSERT OR IGNORE INTO tags (id, user_id, name) VALUES (?, ?, ?)`, idgen.New(), uid, name); err != nil {
				return err
			}
			var tagID string
			if err := tx.QueryRowContext(ctx,
				`SELECT id FROM tags WHERE user_id = ? AND name = ?`, uid, name).Scan(&tagID); err != nil {
				return err
			}
			if _, err := tx.ExecContext(ctx,
				`INSERT OR IGNORE INTO task_tags (task_id, tag_id, user_id) VALUES (?, ?, ?)`,
				string(taskID), tagID, uid); err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *sqliteRepo) listTags(ctx context.Context, taskID ID) ([]string, error) {
	uid, _ := tenant.UserID(ctx)
	rows, err := r.db.QueryContext(ctx,
		`SELECT t.name FROM tags t
		 JOIN task_tags tt ON tt.tag_id = t.id
		 WHERE tt.task_id = ? AND tt.user_id = ? ORDER BY t.name`, string(taskID), uid)
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
