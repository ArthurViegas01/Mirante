-- +goose Up
-- Projetos: a espinha. Projects + links + a shared tag vocabulary (also used by
-- tasks from F2). Times are ISO-8601 UTC text.

CREATE TABLE projects (
    id           TEXT PRIMARY KEY,
    nome         TEXT NOT NULL,
    codinome     TEXT COLLATE NOCASE,
    descricao    TEXT,
    repo         TEXT,
    status       TEXT NOT NULL DEFAULT 'ideia'
                 CHECK (status IN ('ideia','ativo','pausado','no_ar','arquivado')),
    visibilidade TEXT NOT NULL DEFAULT 'pessoal'
                 CHECK (visibilidade IN ('pessoal','lumni','cliente')),
    created_at   TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at   TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);
CREATE UNIQUE INDEX ux_projects_codinome ON projects (codinome) WHERE codinome IS NOT NULL;
CREATE INDEX ix_projects_status ON projects (status);

CREATE TABLE project_links (
    id         TEXT PRIMARY KEY,
    project_id TEXT NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    label      TEXT NOT NULL,
    url        TEXT NOT NULL,
    kind       TEXT NOT NULL DEFAULT 'other'
               CHECK (kind IN ('prod','staging','repo','docs','design','other')),
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);
CREATE INDEX ix_project_links_project ON project_links (project_id, sort_order);

CREATE TABLE tags (
    id   TEXT PRIMARY KEY,
    name TEXT NOT NULL COLLATE NOCASE
);
CREATE UNIQUE INDEX ux_tags_name ON tags (name);

CREATE TABLE project_tags (
    project_id TEXT NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    tag_id     TEXT NOT NULL REFERENCES tags (id) ON DELETE CASCADE,
    PRIMARY KEY (project_id, tag_id)
);
CREATE INDEX ix_project_tags_tag ON project_tags (tag_id);

-- +goose Down
DROP INDEX IF EXISTS ix_project_tags_tag;
DROP TABLE IF EXISTS project_tags;
DROP INDEX IF EXISTS ux_tags_name;
DROP TABLE IF EXISTS tags;
DROP INDEX IF EXISTS ix_project_links_project;
DROP TABLE IF EXISTS project_links;
DROP INDEX IF EXISTS ix_projects_status;
DROP INDEX IF EXISTS ux_projects_codinome;
DROP TABLE IF EXISTS projects;
