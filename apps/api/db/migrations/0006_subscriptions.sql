-- +goose Up
-- Assinaturas: recurring costs attached to a project. A subscription may point at
-- a monitor service (service_id) to show cost beside live status, but there is NO
-- FK to services — they are separate domains (ADR-0001/0005). The only spine FK is
-- project_id → projects ON DELETE CASCADE. Money is an integer in the currency's
-- minor unit (valor_cents); totals are summed per currency, no conversion.

CREATE TABLE subscriptions (
    id          TEXT PRIMARY KEY,
    project_id  TEXT NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    service_id  TEXT,
    nome        TEXT NOT NULL,
    provider    TEXT,
    valor_cents INTEGER NOT NULL DEFAULT 0,
    moeda       TEXT NOT NULL DEFAULT 'BRL' CHECK (moeda IN ('BRL','USD')),
    ciclo       TEXT NOT NULL DEFAULT 'mensal' CHECK (ciclo IN ('mensal','anual')),
    ativo       INTEGER NOT NULL DEFAULT 1 CHECK (ativo IN (0,1)),
    notas       TEXT,
    created_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);
CREATE INDEX ix_subscriptions_project ON subscriptions (project_id);

-- +goose Down
DROP INDEX IF EXISTS ix_subscriptions_project;
DROP TABLE IF EXISTS subscriptions;
