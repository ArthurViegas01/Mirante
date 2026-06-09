-- +goose Up
-- Stack labels for monitored services: provider (free text, e.g. "netlify",
-- "railway") and camada (frontend/backend/database/outro). Both optional; the
-- camada enum is enforced in the Go service layer (validate.Var). The project
-- view groups a project's services by camada and shows live status. See ADR-0005.
ALTER TABLE services ADD COLUMN provider TEXT;
ALTER TABLE services ADD COLUMN camada TEXT;

-- +goose Down
ALTER TABLE services DROP COLUMN camada;
ALTER TABLE services DROP COLUMN provider;
