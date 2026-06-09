-- +goose Up
-- Candidaturas (CRM): the owner's job applications. References a vaga by job_id
-- (soft link, NO FK — applications and jobs are separate domains, ADR-0001), with
-- a snapshot of titulo/empresa so a candidatura survives the vaga being deleted.
-- Pipeline status + a single next-action follow-up (note + date).
CREATE TABLE applications (
    id           TEXT PRIMARY KEY,
    job_id       TEXT,
    titulo       TEXT,
    empresa      TEXT,
    status       TEXT NOT NULL DEFAULT 'interesse'
                 CHECK (status IN ('interesse','aplicado','entrevista','oferta','aceito','rejeitado')),
    notas        TEXT,
    proxima_acao TEXT,
    data_acao    TEXT,
    created_at   TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at   TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);
CREATE INDEX ix_applications_status ON applications (status);

-- +goose Down
DROP INDEX IF EXISTS ix_applications_status;
DROP TABLE IF EXISTS applications;
