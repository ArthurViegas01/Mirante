-- +goose Up
-- Multi-user roles + account activation. The first account (env-bootstrapped or
-- the very first signup) is the admin and is active; everyone who signs up after
-- that starts 'pending' and cannot log in until an admin activates them. status
-- defaults to 'pending' (fail-closed): creation paths set 'active' explicitly.
ALTER TABLE users ADD COLUMN role TEXT NOT NULL DEFAULT 'user'
    CHECK (role IN ('admin', 'user'));
ALTER TABLE users ADD COLUMN status TEXT NOT NULL DEFAULT 'pending'
    CHECK (status IN ('pending', 'active', 'disabled'));

-- The existing single owner becomes the admin and stays active.
UPDATE users SET role = 'admin', status = 'active';

-- +goose Down
ALTER TABLE users DROP COLUMN status;
ALTER TABLE users DROP COLUMN role;
