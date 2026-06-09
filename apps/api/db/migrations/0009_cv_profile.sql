-- +goose Up
-- CV: the owner's master profile (a singleton — single-user app). For now it holds
-- the identity/headline used across the career-search area (e.g. the target role
-- shown on the Vagas header). The full master CV (experiences, education, skills)
-- grows from here in later F3 work.
CREATE TABLE cv_profile (
    id          TEXT PRIMARY KEY,
    nome        TEXT,
    titulo      TEXT,
    titulo_alvo TEXT,
    resumo      TEXT,
    created_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

-- +goose Down
DROP TABLE IF EXISTS cv_profile;
