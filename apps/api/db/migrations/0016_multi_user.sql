-- +goose Up
-- Multi-user: give every domain row an owner (user_id) so the API can isolate
-- data per user. Existing rows are backfilled to the single current owner (the
-- one row in `users`); on a fresh instance the tables are empty and the backfill
-- is a no-op. Globally-unique vocab (project codinome, tags, cv skills) becomes
-- unique PER USER.
--
-- user_id is a plain column (no FK): libSQL runs with foreign_keys off, so a FK
-- would be unenforced in prod anyway. Referential cleanup on user deletion is
-- done explicitly in the app. High-volume monitor history (check_results,
-- check_rollups) stays unscoped here and is isolated via its parent service
-- (services.user_id) in queries.

-- ---- Projects ------------------------------------------------------------
ALTER TABLE projects ADD COLUMN user_id TEXT;
UPDATE projects SET user_id = (SELECT id FROM users LIMIT 1);
CREATE INDEX ix_projects_user ON projects (user_id);
DROP INDEX ux_projects_codinome;
CREATE UNIQUE INDEX ux_projects_user_codinome ON projects (user_id, codinome) WHERE codinome IS NOT NULL;

ALTER TABLE project_links ADD COLUMN user_id TEXT;
UPDATE project_links SET user_id = (SELECT id FROM users LIMIT 1);
CREATE INDEX ix_project_links_user ON project_links (user_id);

ALTER TABLE project_tags ADD COLUMN user_id TEXT;
UPDATE project_tags SET user_id = (SELECT id FROM users LIMIT 1);
CREATE INDEX ix_project_tags_user ON project_tags (user_id);

-- ---- Tags (shared vocabulary → per user) --------------------------------
ALTER TABLE tags ADD COLUMN user_id TEXT;
UPDATE tags SET user_id = (SELECT id FROM users LIMIT 1);
DROP INDEX ux_tags_name;
CREATE UNIQUE INDEX ux_tags_user_name ON tags (user_id, name);

-- ---- Tasks ---------------------------------------------------------------
ALTER TABLE tasks ADD COLUMN user_id TEXT;
UPDATE tasks SET user_id = (SELECT id FROM users LIMIT 1);
CREATE INDEX ix_tasks_user ON tasks (user_id);

ALTER TABLE task_tags ADD COLUMN user_id TEXT;
UPDATE task_tags SET user_id = (SELECT id FROM users LIMIT 1);
CREATE INDEX ix_task_tags_user ON task_tags (user_id);

-- ---- Subscriptions -------------------------------------------------------
ALTER TABLE subscriptions ADD COLUMN user_id TEXT;
UPDATE subscriptions SET user_id = (SELECT id FROM users LIMIT 1);
CREATE INDEX ix_subscriptions_user ON subscriptions (user_id);

-- ---- Jobs ----------------------------------------------------------------
ALTER TABLE jobs ADD COLUMN user_id TEXT;
UPDATE jobs SET user_id = (SELECT id FROM users LIMIT 1);
CREATE INDEX ix_jobs_user ON jobs (user_id);

ALTER TABLE job_skills ADD COLUMN user_id TEXT;
UPDATE job_skills SET user_id = (SELECT id FROM users LIMIT 1);
CREATE INDEX ix_job_skills_user ON job_skills (user_id);

-- ---- Applications --------------------------------------------------------
ALTER TABLE applications ADD COLUMN user_id TEXT;
UPDATE applications SET user_id = (SELECT id FROM users LIMIT 1);
CREATE INDEX ix_applications_user ON applications (user_id);

-- ---- CV (was a single-user singleton → one per user) --------------------
ALTER TABLE cv_profile ADD COLUMN user_id TEXT;
UPDATE cv_profile SET user_id = (SELECT id FROM users LIMIT 1);
CREATE UNIQUE INDEX ux_cv_profile_user ON cv_profile (user_id);

ALTER TABLE cv_experience ADD COLUMN user_id TEXT;
UPDATE cv_experience SET user_id = (SELECT id FROM users LIMIT 1);
CREATE INDEX ix_cv_experience_user ON cv_experience (user_id);

ALTER TABLE cv_education ADD COLUMN user_id TEXT;
UPDATE cv_education SET user_id = (SELECT id FROM users LIMIT 1);
CREATE INDEX ix_cv_education_user ON cv_education (user_id);

-- cv_skills had a global PRIMARY KEY (skill); rebuild so the same skill can
-- belong to several users (PK becomes user_id + skill).
CREATE TABLE cv_skills_new (
    user_id TEXT NOT NULL,
    skill   TEXT NOT NULL,
    PRIMARY KEY (user_id, skill)
);
INSERT INTO cv_skills_new (user_id, skill)
    SELECT (SELECT id FROM users LIMIT 1), skill FROM cv_skills
    WHERE (SELECT id FROM users LIMIT 1) IS NOT NULL;
DROP TABLE cv_skills;
ALTER TABLE cv_skills_new RENAME TO cv_skills;

-- ---- Monitor (services owned; history isolated via the service) ----------
ALTER TABLE services ADD COLUMN user_id TEXT;
UPDATE services SET user_id = (SELECT id FROM users LIMIT 1);
CREATE INDEX ix_services_user ON services (user_id);

ALTER TABLE alerts ADD COLUMN user_id TEXT;
UPDATE alerts SET user_id = (SELECT id FROM users LIMIT 1);
CREATE INDEX ix_alerts_user ON alerts (user_id);

-- SSE outbox: tag each event with its owner so the hub fans out per user.
ALTER TABLE events ADD COLUMN user_id TEXT;
UPDATE events SET user_id = (SELECT id FROM users LIMIT 1);
CREATE INDEX ix_events_user ON events (user_id);

-- ---- LLM usage (per-user accounting) ------------------------------------
ALTER TABLE llm_usage ADD COLUMN user_id TEXT;
UPDATE llm_usage SET user_id = (SELECT id FROM users LIMIT 1);
CREATE INDEX ix_llm_usage_user ON llm_usage (user_id);

-- +goose Down
DROP INDEX IF EXISTS ix_llm_usage_user;
ALTER TABLE llm_usage DROP COLUMN user_id;

DROP INDEX IF EXISTS ix_events_user;
ALTER TABLE events DROP COLUMN user_id;

DROP INDEX IF EXISTS ix_alerts_user;
ALTER TABLE alerts DROP COLUMN user_id;

DROP INDEX IF EXISTS ix_services_user;
ALTER TABLE services DROP COLUMN user_id;

CREATE TABLE cv_skills_old (skill TEXT PRIMARY KEY);
INSERT OR IGNORE INTO cv_skills_old (skill) SELECT skill FROM cv_skills;
DROP TABLE cv_skills;
ALTER TABLE cv_skills_old RENAME TO cv_skills;

DROP INDEX IF EXISTS ix_cv_education_user;
ALTER TABLE cv_education DROP COLUMN user_id;
DROP INDEX IF EXISTS ix_cv_experience_user;
ALTER TABLE cv_experience DROP COLUMN user_id;
DROP INDEX IF EXISTS ux_cv_profile_user;
ALTER TABLE cv_profile DROP COLUMN user_id;

DROP INDEX IF EXISTS ix_applications_user;
ALTER TABLE applications DROP COLUMN user_id;

DROP INDEX IF EXISTS ix_job_skills_user;
ALTER TABLE job_skills DROP COLUMN user_id;
DROP INDEX IF EXISTS ix_jobs_user;
ALTER TABLE jobs DROP COLUMN user_id;

DROP INDEX IF EXISTS ix_subscriptions_user;
ALTER TABLE subscriptions DROP COLUMN user_id;

DROP INDEX IF EXISTS ix_task_tags_user;
ALTER TABLE task_tags DROP COLUMN user_id;
DROP INDEX IF EXISTS ix_tasks_user;
ALTER TABLE tasks DROP COLUMN user_id;

DROP INDEX IF EXISTS ux_tags_user_name;
ALTER TABLE tags DROP COLUMN user_id;
CREATE UNIQUE INDEX ux_tags_name ON tags (name);

DROP INDEX IF EXISTS ix_project_tags_user;
ALTER TABLE project_tags DROP COLUMN user_id;
DROP INDEX IF EXISTS ix_project_links_user;
ALTER TABLE project_links DROP COLUMN user_id;
DROP INDEX IF EXISTS ux_projects_user_codinome;
DROP INDEX IF EXISTS ix_projects_user;
ALTER TABLE projects DROP COLUMN user_id;
CREATE UNIQUE INDEX ux_projects_codinome ON projects (codinome) WHERE codinome IS NOT NULL;
