-- +goose Up
-- The owner's master skills (single-user app, so a flat set). Canonical names
-- (via internal/skills.Normalize) when recognized, otherwise the raw label. These
-- back the deterministic aderência (overlap with a vaga's required skills).
CREATE TABLE cv_skills (
    skill TEXT PRIMARY KEY
);

-- +goose Down
DROP TABLE IF EXISTS cv_skills;
