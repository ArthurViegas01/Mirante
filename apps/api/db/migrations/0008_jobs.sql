-- +goose Up
-- Vagas: job postings the owner is tracking. Skills required by a posting are
-- extracted deterministically from its text (internal/skills) and stored in
-- job_skills (canonical names, not a FK — skills is an in-code kernel). A job is
-- standalone (no project FK); tasks may reference a job by id (soft link).
CREATE TABLE jobs (
    id          TEXT PRIMARY KEY,
    titulo      TEXT NOT NULL,
    empresa     TEXT,
    descricao   TEXT,
    url         TEXT,
    localizacao TEXT,
    modelo      TEXT NOT NULL DEFAULT 'indefinido'
                CHECK (modelo IN ('remoto','hibrido','presencial','indefinido')),
    senioridade TEXT,
    resumo      TEXT,
    created_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);
CREATE INDEX ix_jobs_created ON jobs (created_at);

CREATE TABLE job_skills (
    job_id TEXT NOT NULL REFERENCES jobs (id) ON DELETE CASCADE,
    skill  TEXT NOT NULL,
    PRIMARY KEY (job_id, skill)
);

-- +goose Down
DROP TABLE IF EXISTS job_skills;
DROP INDEX IF EXISTS ix_jobs_created;
DROP TABLE IF EXISTS jobs;
