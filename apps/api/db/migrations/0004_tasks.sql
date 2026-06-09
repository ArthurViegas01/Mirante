-- +goose Up
-- Tarefas: kanban work items. A task can hang off a project (FK, ON DELETE SET
-- NULL so finished work outlives an archived or removed project) and, from F3, a
-- job (job_id is reserved now; its FK arrives with the jobs domain). Tags reuse
-- the shared vocabulary from 0002. Times are ISO-8601 UTC text; prazo is a
-- calendar date (YYYY-MM-DD).

CREATE TABLE tasks (
    id         TEXT PRIMARY KEY,
    titulo     TEXT NOT NULL,
    descricao  TEXT,
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

CREATE TABLE task_tags (
    task_id TEXT NOT NULL REFERENCES tasks (id) ON DELETE CASCADE,
    tag_id  TEXT NOT NULL REFERENCES tags (id) ON DELETE CASCADE,
    PRIMARY KEY (task_id, tag_id)
);
CREATE INDEX ix_task_tags_tag ON task_tags (tag_id);

-- +goose Down
DROP INDEX IF EXISTS ix_task_tags_tag;
DROP TABLE IF EXISTS task_tags;
DROP INDEX IF EXISTS ix_tasks_project;
DROP INDEX IF EXISTS ix_tasks_status;
DROP TABLE IF EXISTS tasks;
