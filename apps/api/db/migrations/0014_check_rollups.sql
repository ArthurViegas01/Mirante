-- +goose Up
-- Monitor F4: hourly rollups of check_results so the raw time series can be
-- pruned past a retention window while long-range uptime stays computable.
-- One row per (service, hour bucket); a bucket is 'YYYY-MM-DDTHH' (UTC), matching
-- substr(checked_at,1,13) of the canonical ISO timestamp. The compactor sums raw
-- rows older than retention into these buckets, then deletes the raw rows — so
-- rollups and the surviving raw rows are disjoint in time (no double counting).

CREATE TABLE check_rollups (
    service_id     TEXT NOT NULL REFERENCES services (id) ON DELETE CASCADE,
    bucket         TEXT NOT NULL, -- 'YYYY-MM-DDTHH' UTC hour
    samples        INTEGER NOT NULL,
    ups            INTEGER NOT NULL, -- count of outcome != 'down' (up + degraded)
    sum_latency_ms INTEGER NOT NULL, -- reserved for long-window average latency
    PRIMARY KEY (service_id, bucket)
);

-- +goose Down
DROP TABLE IF EXISTS check_rollups;
