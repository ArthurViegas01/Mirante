-- +goose Up
-- Single-use password-reset tokens for the owner's "forgot password" flow.
-- The e-mail carries the plaintext token; only its SHA-256 is stored here, so a
-- database leak cannot redeem a reset (same posture as sessions). Times are
-- ISO-8601 UTC text. A token is valid until expires_at and at most once (used_at).

CREATE TABLE password_resets (
    id         TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    expires_at TEXT NOT NULL,
    used_at    TEXT
);

CREATE UNIQUE INDEX ux_password_resets_token_hash ON password_resets (token_hash);
CREATE INDEX ix_password_resets_user_id ON password_resets (user_id);
CREATE INDEX ix_password_resets_expires_at ON password_resets (expires_at);

-- +goose Down
DROP INDEX IF EXISTS ix_password_resets_expires_at;
DROP INDEX IF EXISTS ix_password_resets_user_id;
DROP INDEX IF EXISTS ux_password_resets_token_hash;
DROP TABLE IF EXISTS password_resets;
