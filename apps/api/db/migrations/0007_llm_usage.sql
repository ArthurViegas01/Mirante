-- +goose Up
-- LLM usage ledger (ADR-0004): one row per model call, for cost/audit and to back
-- per-route rate limiting. Tokens are provider-reported; monetary cost is derived
-- off-line from provider price tables, so it is not stored here.
CREATE TABLE llm_usage (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    provider      TEXT NOT NULL,
    model         TEXT NOT NULL,
    route         TEXT NOT NULL,
    input_tokens  INTEGER NOT NULL DEFAULT 0,
    output_tokens INTEGER NOT NULL DEFAULT 0,
    created_at    TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);
CREATE INDEX ix_llm_usage_created ON llm_usage (created_at);

-- +goose Down
DROP INDEX IF EXISTS ix_llm_usage_created;
DROP TABLE IF EXISTS llm_usage;
