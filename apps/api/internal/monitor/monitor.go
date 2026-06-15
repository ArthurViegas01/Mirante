// Package monitor watches a project's services, derives up/degraded/down state
// with anti-flap, records history, raises in-app alerts, and emits live events.
package monitor

import (
	"encoding/json"
	"time"
)

// ServiceID identifies a monitored service.
type ServiceID string

// Kind is the check type.
type Kind string

const (
	KindHTTP   Kind = "http"
	KindTCP    Kind = "tcp"
	KindDBPing Kind = "db_ping"
)

// Status is the derived health state.
type Status string

const (
	StatusUnknown  Status = "unknown"
	StatusUp       Status = "up"
	StatusDegraded Status = "degraded"
	StatusDown     Status = "down"
	StatusPaused   Status = "paused"
)

// Service is a monitored endpoint belonging to a project.
type Service struct {
	ID                   ServiceID  `json:"id"`
	UserID               string     `json:"-"` // owner; set on read, stamped on write
	ProjectID            string     `json:"project_id"`
	Nome                 string     `json:"nome"`
	Provider             string     `json:"provider"` // free label, e.g. "netlify"
	Camada               string     `json:"camada"`   // frontend|backend|database|outro
	Kind                 Kind       `json:"kind"`
	Target               string     `json:"target"`
	ExpectedStatus       string     `json:"expected_status"`
	DegradedThresholdMs  int        `json:"degraded_threshold_ms"`
	TimeoutMs            int        `json:"timeout_ms"`
	IntervalSeconds      int        `json:"interval_seconds"`
	AntiFlapN            int        `json:"anti_flap_n"`
	RecoveryK            int        `json:"recovery_k"`
	Enabled              bool       `json:"enabled"`
	CurrentStatus        Status     `json:"current_status"`
	ConsecutiveFailures  int        `json:"-"`
	ConsecutiveSuccesses int        `json:"-"`
	LastCheckedAt        *time.Time `json:"last_checked_at"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

// CheckResult is one probe stored in the time series.
type CheckResult struct {
	ID         int64     `json:"id"`
	ServiceID  ServiceID `json:"service_id"`
	CheckedAt  time.Time `json:"checked_at"`
	OK         bool      `json:"ok"`
	Outcome    Status    `json:"outcome"`
	LatencyMs  int       `json:"latency_ms"`
	StatusCode int       `json:"status_code"`
	ErrorKind  string    `json:"error_kind"`
}

// Alert is an in-app notification raised on a status transition.
type Alert struct {
	ID         int64      `json:"id"`
	ServiceID  ServiceID  `json:"service_id"`
	ProjectID  string     `json:"project_id"`
	Severity   string     `json:"severity"`
	Title      string     `json:"title"`
	Body       string     `json:"body"`
	FromStatus Status     `json:"from_status"`
	ToStatus   Status     `json:"to_status"`
	ReadAt     *time.Time `json:"read_at"`
	CreatedAt  time.Time  `json:"created_at"`
}

// Event is a row in the SSE outbox; its ID is the durable Last-Event-ID.
type Event struct {
	ID        int64           `json:"id"`
	Type      string          `json:"type"`
	Data      json.RawMessage `json:"data"`
	CreatedAt time.Time       `json:"created_at"`
}

// Uptime is the rolling availability over a window.
type Uptime struct {
	WindowHours int     `json:"window_hours"`
	Samples     int     `json:"samples"`
	UpRatio     float64 `json:"up_ratio"` // (up+degraded) / samples
}

// severityFor maps a target status to a semantic severity (design tokens).
func severityFor(to Status) string {
	switch to {
	case StatusUp:
		return "success"
	case StatusDegraded:
		return "warning"
	case StatusDown:
		return "danger"
	default:
		return "info"
	}
}
