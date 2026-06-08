package monitor

import (
	"context"
	"errors"
	"time"
)

// ErrNotFound is returned when a service does not exist.
var ErrNotFound = errors.New("service not found")

// RecordCheckInput is the data persisted after one probe.
type RecordCheckInput struct {
	Service    *Service
	From       Status
	Result     DeriveResult
	LatencyMs  int
	StatusCode int
	ErrorKind  string
	CheckedAt  time.Time
}

// RecordCheckOutput carries the alert and event created on a transition.
type RecordCheckOutput struct {
	Alert *Alert
	Event *Event
}

// Repository persists monitor state.
type Repository interface {
	CreateService(ctx context.Context, s *Service) error
	GetService(ctx context.Context, id ServiceID) (*Service, error)
	ListServices(ctx context.Context, projectID string) ([]*Service, error) // projectID "" = all
	ListEnabledServices(ctx context.Context) ([]*Service, error)
	CountServicesByProject(ctx context.Context, projectID string) (int, error)
	UpdateServiceConfig(ctx context.Context, s *Service) error
	SetServiceStatus(ctx context.Context, id ServiceID, status Status, resetCounters bool) error
	DeleteService(ctx context.Context, id ServiceID) error

	// RecordCheck atomically writes the check result, updates the service's
	// status/counters, and (on a transition) inserts the alert and the SSE
	// outbox event, returning them.
	RecordCheck(ctx context.Context, in RecordCheckInput) (RecordCheckOutput, error)

	ListChecks(ctx context.Context, id ServiceID, limit int) ([]CheckResult, error)
	Uptime(ctx context.Context, id ServiceID, windowHours int) (Uptime, error)

	ListAlerts(ctx context.Context, limit int, unreadOnly bool) ([]Alert, error)
	CountUnreadAlerts(ctx context.Context) (int, error)
	MarkAlertRead(ctx context.Context, id int64) error
	MarkAllAlertsRead(ctx context.Context) error

	EventsAfter(ctx context.Context, afterID int64, limit int) ([]Event, error)
}
