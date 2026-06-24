-- +goose Up
-- Intake: staging area for freelance opportunities pulled from external feeds
-- (today: 99Freelas digest e-mails). High-volume and low-commitment — most items
-- are triaged and discarded; only the chosen few are promoted out of staging into
-- a tracked vaga (jobs) + candidatura. Scoped per user (the owner of the inbox the
-- poller reads). Dedup is by (user_id, fonte, fonte_id): the same project recurs
-- across daily digests. user_id is a plain column (no FK), matching the schema.
CREATE TABLE intake_items (
    id             TEXT PRIMARY KEY,
    user_id        TEXT NOT NULL,
    fonte          TEXT NOT NULL,
    fonte_id       TEXT NOT NULL,
    titulo         TEXT NOT NULL,
    categoria      TEXT,
    nivel          TEXT,
    publicado      TEXT,
    tempo_restante TEXT,
    restante_horas INTEGER NOT NULL DEFAULT 0,
    propostas      INTEGER NOT NULL DEFAULT 0,
    interessados   INTEGER NOT NULL DEFAULT 0,
    teaser         TEXT,
    url            TEXT,
    enviar_url     TEXT,
    skills         TEXT,
    score          INTEGER NOT NULL DEFAULT 0,
    estado         TEXT NOT NULL DEFAULT 'novo'
                   CHECK (estado IN ('novo','descartado','promovido')),
    created_at     TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at     TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);
CREATE UNIQUE INDEX ux_intake_items_dedup ON intake_items (user_id, fonte, fonte_id);
CREATE INDEX ix_intake_items_user_score ON intake_items (user_id, score);

-- +goose Down
DROP INDEX IF EXISTS ix_intake_items_user_score;
DROP INDEX IF EXISTS ux_intake_items_dedup;
DROP TABLE IF EXISTS intake_items;
