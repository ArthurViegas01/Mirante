-- +goose Up
-- Monitor: services + time-series checks + in-app alerts + a single SSE event
-- outbox (its autoincrement id is the durable Last-Event-ID sequence).

CREATE TABLE services (
    id                    TEXT PRIMARY KEY,
    project_id            TEXT NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    nome                  TEXT NOT NULL,
    kind                  TEXT NOT NULL CHECK (kind IN ('http','tcp','db_ping')),
    target                TEXT NOT NULL,
    expected_status       TEXT NOT NULL DEFAULT '2xx', -- http: '2xx' | '200,204'
    degraded_threshold_ms INTEGER NOT NULL DEFAULT 500,
    timeout_ms            INTEGER NOT NULL DEFAULT 5000,
    interval_seconds      INTEGER NOT NULL DEFAULT 60,
    anti_flap_n           INTEGER NOT NULL DEFAULT 3,
    recovery_k            INTEGER NOT NULL DEFAULT 2,
    enabled               INTEGER NOT NULL DEFAULT 1 CHECK (enabled IN (0,1)),
    current_status        TEXT NOT NULL DEFAULT 'unknown'
                          CHECK (current_status IN ('unknown','up','degraded','down','paused')),
    consecutive_failures  INTEGER NOT NULL DEFAULT 0,
    consecutive_successes INTEGER NOT NULL DEFAULT 0,
    last_checked_at       TEXT,
    created_at            TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at            TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    CHECK (interval_seconds >= 5),
    CHECK (timeout_ms < interval_seconds * 1000)
);
CREATE INDEX ix_services_project ON services (project_id);
CREATE INDEX ix_services_enabled ON services (enabled);

CREATE TABLE check_results (
    id          INTEGER PRIMARY KEY,
    service_id  TEXT NOT NULL REFERENCES services (id) ON DELETE CASCADE,
    checked_at  TEXT NOT NULL,
    ok          INTEGER NOT NULL CHECK (ok IN (0,1)),
    outcome     TEXT NOT NULL CHECK (outcome IN ('up','degraded','down')),
    latency_ms  INTEGER,
    status_code INTEGER,
    error_kind  TEXT
);
CREATE INDEX ix_check_results_service_time ON check_results (service_id, checked_at);

CREATE TABLE alerts (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    service_id  TEXT NOT NULL REFERENCES services (id) ON DELETE CASCADE,
    project_id  TEXT NOT NULL,
    severity    TEXT NOT NULL CHECK (severity IN ('success','warning','danger','info')),
    title       TEXT NOT NULL,
    body        TEXT,
    from_status TEXT,
    to_status   TEXT,
    read_at     TEXT,
    created_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);
CREATE INDEX ix_alerts_created ON alerts (created_at);
CREATE INDEX ix_alerts_unread ON alerts (read_at) WHERE read_at IS NULL;

CREATE TABLE events (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    type       TEXT NOT NULL,
    data       TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

-- +goose Down
DROP TABLE IF EXISTS events;
DROP INDEX IF EXISTS ix_alerts_unread;
DROP INDEX IF EXISTS ix_alerts_created;
DROP TABLE IF EXISTS alerts;
DROP INDEX IF EXISTS ix_check_results_service_time;
DROP TABLE IF EXISTS check_results;
DROP INDEX IF EXISTS ix_services_enabled;
DROP INDEX IF EXISTS ix_services_project;
DROP TABLE IF EXISTS services;
