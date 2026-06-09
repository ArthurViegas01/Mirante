-- +goose Up
-- Contact line for the master CV (email · phone · location · links), shown on the
-- exported PDF/DOCX. A single free-text field keeps it simple; the LLM import
-- fills it from the pasted CV's header.
ALTER TABLE cv_profile ADD COLUMN contato TEXT;

-- +goose Down
ALTER TABLE cv_profile DROP COLUMN contato;
