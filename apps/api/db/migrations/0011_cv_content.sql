-- +goose Up
-- Master CV content: experiences and education (single-user, so flat lists with a
-- display order). Fully replaced on each save (PUT /api/cv), so ids are server-
-- assigned. Dates are free text ("YYYY-MM" or "2023", "atual"…) — a CV is prose.
CREATE TABLE cv_experience (
    id        TEXT PRIMARY KEY,
    empresa   TEXT,
    cargo     TEXT,
    inicio    TEXT,
    fim       TEXT,
    descricao TEXT,
    ordem     INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX ix_cv_experience_ordem ON cv_experience (ordem);

CREATE TABLE cv_education (
    id          TEXT PRIMARY KEY,
    instituicao TEXT,
    curso       TEXT,
    inicio      TEXT,
    fim         TEXT,
    ordem       INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX ix_cv_education_ordem ON cv_education (ordem);

-- +goose Down
DROP INDEX IF EXISTS ix_cv_education_ordem;
DROP TABLE IF EXISTS cv_education;
DROP INDEX IF EXISTS ix_cv_experience_ordem;
DROP TABLE IF EXISTS cv_experience;
