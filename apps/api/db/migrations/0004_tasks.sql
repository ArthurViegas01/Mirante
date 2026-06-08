-- +goose Up
-- Tarefas: activities linkable to a project and (from F3) to a job. Reuses the
-- shared `tags` vocabulary. project_id is a nullable FK→projects with ON DELETE
-- SET NULL so deleting a project detaches its tasks instead of destroying them
-- (a task is legitimately project-less). job_id is nullable with NO FK yet — the
-- constraint FK→jobs lands in an F3 migration. Times are ISO-8601 UTC text.

CREATE TABLE tasks (
    id         TEXT PRIMARY KEY,
    titulo     TEXT NOT NULL,
    status     TEXT NOT NULL DEFAULT 'a_fazer'
               CHECK (status IN ('a_fazer','fazendo','feito')),
    prioridade TEXT NOT NULL DEFAULT 'media'
               CHECK (prioridade IN ('baixa','media','alta')),
    prazo      TEXT,
    project_id TEXT REFERENCES projects (id) ON DELETE SET NULL,
    job_id     TEXT,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);
CREATE INDEX ix_tasks_status ON tasks (status);
CREATE INDEX ix_tasks_project ON tasks (project_id);
CREATE INDEX ix_tasks_prazo ON tasks (prazo) WHERE prazo IS NOT NULL;

CREATE TABLE task_tags (
    task_id TEXT NOT NULL REFERENCES tasks (id) ON DELETE CASCADE,
    tag_id  TEXT NOT NULL REFERENCES tags (id) ON DELETE CASCADE,
    PRIMARY KEY (task_id, tag_id)
);
CREATE INDEX ix_task_tags_tag ON task_tags (tag_id);

-- +goose Down
DROP INDEX IF EXISTS ix_task_tags_tag;
DROP TABLE IF EXISTS task_tags;
DROP INDEX IF EXISTS ix_tasks_prazo;
DROP INDEX IF EXISTS ix_tasks_project;
DROP INDEX IF EXISTS ix_tasks_status;
DROP TABLE IF EXISTS tasks;
