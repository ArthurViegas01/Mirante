package monitor

import (
	"context"
	"database/sql"
	"errors"
	"time"

	idb "github.com/lumni/mirante/internal/platform/db"
	"github.com/lumni/mirante/internal/platform/tenant"
)

type sqliteRepo struct{ db *idb.DB }

// NewSQLiteRepo builds a SQLite-backed monitor repository.
func NewSQLiteRepo(d *idb.DB) Repository { return &sqliteRepo{db: d} }

type rowScanner interface{ Scan(dest ...any) error }

// Scoping note: services/alerts/events carry user_id and the API methods filter
// by the request's owner (tenant). The scheduler/engine/compactor run with no
// tenant and use the explicitly unscoped methods (ListEnabledServices,
// RecordCheck, Compact). check_results/check_rollups have no user_id and are
// isolated via their parent service in read queries.
const serviceCols = `id, user_id, project_id, nome, provider, camada, kind, target, expected_status,
	degraded_threshold_ms, timeout_ms, interval_seconds, anti_flap_n, recovery_k,
	enabled, current_status, consecutive_failures, consecutive_successes,
	last_checked_at, created_at, updated_at`

func scanService(s rowScanner) (*Service, error) {
	var (
		svc                                  Service
		idStr, projectID, nome, kind, target string
		userID, expectedStatus, current      string
		provider, camada, lastChecked        sql.NullString
		enabled                              int
		createdAt, updatedAt                 string
	)
	if err := s.Scan(&idStr, &userID, &projectID, &nome, &provider, &camada, &kind, &target, &expectedStatus,
		&svc.DegradedThresholdMs, &svc.TimeoutMs, &svc.IntervalSeconds, &svc.AntiFlapN, &svc.RecoveryK,
		&enabled, &current, &svc.ConsecutiveFailures, &svc.ConsecutiveSuccesses,
		&lastChecked, &createdAt, &updatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	svc.ID = ServiceID(idStr)
	svc.UserID = userID
	svc.ProjectID = projectID
	svc.Nome = nome
	svc.Provider = provider.String
	svc.Camada = camada.String
	svc.Kind = Kind(kind)
	svc.Target = target
	svc.ExpectedStatus = expectedStatus
	svc.Enabled = enabled == 1
	svc.CurrentStatus = Status(current)
	if lastChecked.Valid {
		t := idb.ParseTime(lastChecked.String)
		svc.LastCheckedAt = &t
	}
	svc.CreatedAt = idb.ParseTime(createdAt)
	svc.UpdatedAt = idb.ParseTime(updatedAt)
	return &svc, nil
}

func (r *sqliteRepo) CreateService(ctx context.Context, s *Service) error {
	uid, _ := tenant.UserID(ctx)
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO services (id, user_id, project_id, nome, provider, camada, kind, target, expected_status,
			degraded_threshold_ms, timeout_ms, interval_seconds, anti_flap_n, recovery_k, enabled)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		string(s.ID), uid, s.ProjectID, s.Nome, nullableStr(s.Provider), nullableStr(s.Camada), string(s.Kind), s.Target, s.ExpectedStatus,
		s.DegradedThresholdMs, s.TimeoutMs, s.IntervalSeconds, s.AntiFlapN, s.RecoveryK, boolToInt(s.Enabled))
	return err
}

func (r *sqliteRepo) GetService(ctx context.Context, id ServiceID) (*Service, error) {
	uid, _ := tenant.UserID(ctx)
	return scanService(r.db.QueryRowContext(ctx,
		`SELECT `+serviceCols+` FROM services WHERE id = ? AND user_id = ?`, string(id), uid))
}

func (r *sqliteRepo) ListServices(ctx context.Context, projectID string) ([]*Service, error) {
	uid, _ := tenant.UserID(ctx)
	query := `SELECT ` + serviceCols + ` FROM services WHERE user_id = ?`
	args := []any{uid}
	if projectID != "" {
		query += ` AND project_id = ?`
		args = append(args, projectID)
	}
	query += ` ORDER BY created_at`
	return r.queryServices(ctx, query, args...)
}

// ListEnabledServices returns every enabled service across all users. It is the
// scheduler's system-level enumeration and is intentionally NOT user-scoped.
func (r *sqliteRepo) ListEnabledServices(ctx context.Context) ([]*Service, error) {
	return r.queryServices(ctx, `SELECT `+serviceCols+` FROM services WHERE enabled = 1`)
}

func (r *sqliteRepo) queryServices(ctx context.Context, query string, args ...any) ([]*Service, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := []*Service{}
	for rows.Next() {
		svc, err := scanService(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, svc)
	}
	return out, rows.Err()
}

func (r *sqliteRepo) CountServicesByProject(ctx context.Context, projectID string) (int, error) {
	uid, _ := tenant.UserID(ctx)
	var n int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM services WHERE project_id = ? AND user_id = ?`, projectID, uid).Scan(&n)
	return n, err
}

func (r *sqliteRepo) UpdateServiceConfig(ctx context.Context, s *Service) error {
	uid, _ := tenant.UserID(ctx)
	res, err := r.db.ExecContext(ctx,
		`UPDATE services SET nome = ?, provider = ?, camada = ?, kind = ?, target = ?, expected_status = ?,
			degraded_threshold_ms = ?, timeout_ms = ?, interval_seconds = ?,
			anti_flap_n = ?, recovery_k = ?, enabled = ?, updated_at = ?
		 WHERE id = ? AND user_id = ?`,
		s.Nome, nullableStr(s.Provider), nullableStr(s.Camada), string(s.Kind), s.Target, s.ExpectedStatus,
		s.DegradedThresholdMs, s.TimeoutMs, s.IntervalSeconds, s.AntiFlapN, s.RecoveryK,
		boolToInt(s.Enabled), idb.FormatTime(time.Now()), string(s.ID), uid)
	if err != nil {
		return err
	}
	return affected(res)
}

func (r *sqliteRepo) SetServiceStatus(ctx context.Context, id ServiceID, status Status, resetCounters bool) error {
	uid, _ := tenant.UserID(ctx)
	if resetCounters {
		_, err := r.db.ExecContext(ctx,
			`UPDATE services SET current_status = ?, consecutive_failures = 0, consecutive_successes = 0
			 WHERE id = ? AND user_id = ?`,
			string(status), string(id), uid)
		return err
	}
	_, err := r.db.ExecContext(ctx,
		`UPDATE services SET current_status = ? WHERE id = ? AND user_id = ?`, string(status), string(id), uid)
	return err
}

func (r *sqliteRepo) DeleteService(ctx context.Context, id ServiceID) error {
	uid, _ := tenant.UserID(ctx)
	res, err := r.db.ExecContext(ctx, `DELETE FROM services WHERE id = ? AND user_id = ?`, string(id), uid)
	if err != nil {
		return err
	}
	return affected(res)
}

// RecordCheck is system-level (the engine, no tenant). It stamps the alert and
// SSE event with the service's owner so they reach only that user.
func (r *sqliteRepo) RecordCheck(ctx context.Context, in RecordCheckInput) (RecordCheckOutput, error) {
	var out RecordCheckOutput
	err := r.db.WithTx(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO check_results (service_id, checked_at, ok, outcome, latency_ms, status_code, error_kind)
			 VALUES (?, ?, ?, ?, ?, ?, ?)`,
			string(in.Service.ID), idb.FormatTime(in.CheckedAt), boolToInt(in.Result.Outcome != StatusDown),
			string(in.Result.Outcome), nullableInt(in.LatencyMs), nullableInt(in.StatusCode), nullableStr(in.ErrorKind),
		); err != nil {
			return err
		}

		if _, err := tx.ExecContext(ctx,
			`UPDATE services SET current_status = ?, consecutive_failures = ?,
				consecutive_successes = ?, last_checked_at = ? WHERE id = ?`,
			string(in.Result.State), in.Result.ConsecFailures, in.Result.ConsecSuccesses,
			idb.FormatTime(in.CheckedAt), string(in.Service.ID),
		); err != nil {
			return err
		}

		if !in.Result.Changed {
			return nil
		}

		alert := buildAlert(in.Service, in.From, in.Result.State, in.Result.Reason)
		alert.CreatedAt = in.CheckedAt
		if err := tx.QueryRowContext(ctx,
			`INSERT INTO alerts (service_id, user_id, project_id, severity, title, body, from_status, to_status, created_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING id`,
			string(alert.ServiceID), in.Service.UserID, alert.ProjectID, alert.Severity, alert.Title, nullableStr(alert.Body),
			string(alert.FromStatus), string(alert.ToStatus), idb.FormatTime(alert.CreatedAt),
		).Scan(&alert.ID); err != nil {
			return err
		}

		data, err := eventData(alert, in.Service.Nome, in.LatencyMs)
		if err != nil {
			return err
		}
		ev := Event{Type: "monitor.transition", Data: data, CreatedAt: in.CheckedAt}
		if err := tx.QueryRowContext(ctx,
			`INSERT INTO events (type, user_id, data, created_at) VALUES (?, ?, ?, ?) RETURNING id`,
			ev.Type, in.Service.UserID, string(data), idb.FormatTime(ev.CreatedAt),
		).Scan(&ev.ID); err != nil {
			return err
		}

		out.Alert = &alert
		out.Event = &ev
		return nil
	})
	return out, err
}

func (r *sqliteRepo) ListChecks(ctx context.Context, id ServiceID, limit int) ([]CheckResult, error) {
	if limit <= 0 {
		limit = 60
	}
	uid, _ := tenant.UserID(ctx)
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, service_id, checked_at, ok, outcome, latency_ms, status_code, error_kind
		 FROM check_results
		 WHERE service_id = ? AND service_id IN (SELECT id FROM services WHERE user_id = ?)
		 ORDER BY id DESC LIMIT ?`, string(id), uid, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	out := []CheckResult{}
	for rows.Next() {
		var (
			c                   CheckResult
			sid, outcome        string
			latency, statusCode sql.NullInt64
			errorKind           sql.NullString
			ok                  int
			checkedAt           string
		)
		if err := rows.Scan(&c.ID, &sid, &checkedAt, &ok, &outcome, &latency, &statusCode, &errorKind); err != nil {
			return nil, err
		}
		c.ServiceID = ServiceID(sid)
		c.CheckedAt = idb.ParseTime(checkedAt)
		c.OK = ok == 1
		c.Outcome = Status(outcome)
		c.LatencyMs = int(latency.Int64)
		c.StatusCode = int(statusCode.Int64)
		c.ErrorKind = errorKind.String
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *sqliteRepo) Uptime(ctx context.Context, id ServiceID, windowHours int) (Uptime, error) {
	uid, _ := tenant.UserID(ctx)
	since := idb.FormatTime(time.Now().UTC().Add(-time.Duration(windowHours) * time.Hour))
	var samples, ups int
	if err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*), COALESCE(SUM(CASE WHEN outcome != 'down' THEN 1 ELSE 0 END), 0)
		 FROM check_results
		 WHERE service_id = ? AND checked_at >= ?
		   AND service_id IN (SELECT id FROM services WHERE user_id = ?)`,
		string(id), since, uid).Scan(&samples, &ups); err != nil {
		return Uptime{}, err
	}

	// Fold in compacted history. Rollup buckets cover raw rows the compactor has
	// deleted, so they are disjoint from the raw rows counted above — summing both
	// cannot double count. since[:13] is the window-start hour ('YYYY-MM-DDTHH'),
	// matching the bucket format (substr of the canonical ISO timestamp).
	var rSamples, rUps int
	if err := r.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(samples), 0), COALESCE(SUM(ups), 0)
		 FROM check_rollups
		 WHERE service_id = ? AND bucket >= ?
		   AND service_id IN (SELECT id FROM services WHERE user_id = ?)`,
		string(id), since[:13], uid).Scan(&rSamples, &rUps); err != nil {
		return Uptime{}, err
	}
	samples += rSamples
	ups += rUps

	u := Uptime{WindowHours: windowHours, Samples: samples}
	if samples > 0 {
		u.UpRatio = float64(ups) / float64(samples)
	}
	return u, nil
}

// Compact is system-level (the retention job, no tenant). It rolls up and prunes
// raw checks across all services; check_results/check_rollups carry no user_id.
func (r *sqliteRepo) Compact(ctx context.Context, before time.Time) (int, error) {
	cutoff := idb.FormatTime(before)
	var pruned int
	err := r.db.WithTx(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO check_rollups (service_id, bucket, samples, ups, sum_latency_ms)
			 SELECT service_id, substr(checked_at, 1, 13),
			        COUNT(*),
			        COALESCE(SUM(CASE WHEN outcome != 'down' THEN 1 ELSE 0 END), 0),
			        COALESCE(SUM(latency_ms), 0)
			 FROM check_results WHERE checked_at < ?
			 GROUP BY service_id, substr(checked_at, 1, 13)
			 ON CONFLICT(service_id, bucket) DO UPDATE SET
			     samples        = samples + excluded.samples,
			     ups            = ups + excluded.ups,
			     sum_latency_ms = sum_latency_ms + excluded.sum_latency_ms`,
			cutoff); err != nil {
			return err
		}
		res, err := tx.ExecContext(ctx, `DELETE FROM check_results WHERE checked_at < ?`, cutoff)
		if err != nil {
			return err
		}
		n, err := res.RowsAffected()
		if err != nil {
			return err
		}
		pruned = int(n)
		return nil
	})
	return pruned, err
}

func (r *sqliteRepo) ListAlerts(ctx context.Context, limit int, unreadOnly bool) ([]Alert, error) {
	if limit <= 0 {
		limit = 50
	}
	uid, _ := tenant.UserID(ctx)
	query := `SELECT id, service_id, project_id, severity, title, body, from_status, to_status, read_at, created_at
		 FROM alerts WHERE user_id = ?`
	if unreadOnly {
		query += ` AND read_at IS NULL`
	}
	query += ` ORDER BY id DESC LIMIT ?`
	rows, err := r.db.QueryContext(ctx, query, uid, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	out := []Alert{}
	for rows.Next() {
		a, err := scanAlert(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func scanAlert(s rowScanner) (Alert, error) {
	var (
		a                        Alert
		sid                      string
		body, fromS, toS, readAt sql.NullString
		createdAt                string
	)
	if err := s.Scan(&a.ID, &sid, &a.ProjectID, &a.Severity, &a.Title, &body, &fromS, &toS, &readAt, &createdAt); err != nil {
		return Alert{}, err
	}
	a.ServiceID = ServiceID(sid)
	a.Body = body.String
	a.FromStatus = Status(fromS.String)
	a.ToStatus = Status(toS.String)
	if readAt.Valid {
		t := idb.ParseTime(readAt.String)
		a.ReadAt = &t
	}
	a.CreatedAt = idb.ParseTime(createdAt)
	return a, nil
}

func (r *sqliteRepo) CountUnreadAlerts(ctx context.Context) (int, error) {
	uid, _ := tenant.UserID(ctx)
	var n int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM alerts WHERE read_at IS NULL AND user_id = ?`, uid).Scan(&n)
	return n, err
}

func (r *sqliteRepo) MarkAlertRead(ctx context.Context, id int64) error {
	uid, _ := tenant.UserID(ctx)
	_, err := r.db.ExecContext(ctx,
		`UPDATE alerts SET read_at = ? WHERE id = ? AND user_id = ? AND read_at IS NULL`,
		idb.FormatTime(time.Now()), id, uid)
	return err
}

func (r *sqliteRepo) MarkAllAlertsRead(ctx context.Context) error {
	uid, _ := tenant.UserID(ctx)
	_, err := r.db.ExecContext(ctx,
		`UPDATE alerts SET read_at = ? WHERE user_id = ? AND read_at IS NULL`, idb.FormatTime(time.Now()), uid)
	return err
}

// EventsAfter replays missed SSE events for the connecting client, scoped to its
// owner (the stream runs behind auth, so the context carries the tenant).
func (r *sqliteRepo) EventsAfter(ctx context.Context, afterID int64, limit int) ([]Event, error) {
	if limit <= 0 {
		limit = 100
	}
	uid, _ := tenant.UserID(ctx)
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, type, data, created_at FROM events
		 WHERE id > ? AND user_id = ? ORDER BY id ASC LIMIT ?`, afterID, uid, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	out := []Event{}
	for rows.Next() {
		var (
			ev        Event
			data      string
			createdAt string
		)
		if err := rows.Scan(&ev.ID, &ev.Type, &data, &createdAt); err != nil {
			return nil, err
		}
		ev.Data = []byte(data)
		ev.CreatedAt = idb.ParseTime(createdAt)
		out = append(out, ev)
	}
	return out, rows.Err()
}

func affected(res sql.Result) error {
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func nullableInt(n int) any {
	if n == 0 {
		return nil
	}
	return n
}

func nullableStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}
