package monitor

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"

	idgen "github.com/lumni/mirante/internal/platform/id"
	"github.com/lumni/mirante/internal/platform/validate"
)

// ErrInvalid wraps validation failures (mapped to HTTP 400).
var ErrInvalid = errors.New("invalid input")

const maxServicesPerProject = 50

// Reconciler lets the manager nudge the scheduler to re-read services promptly
// after a change, so new/edited services start being checked without waiting for
// the periodic reconcile.
type Reconciler interface{ Trigger() }

// Manager holds the monitor's HTTP-facing use cases.
type Manager struct {
	repo       Repository
	reconciler Reconciler
}

// NewManager builds the monitor manager.
func NewManager(repo Repository) *Manager { return &Manager{repo: repo} }

// SetReconciler wires the scheduler so config changes take effect promptly.
func (m *Manager) SetReconciler(r Reconciler) { m.reconciler = r }

func (m *Manager) notifyReconcile() {
	if m.reconciler != nil {
		m.reconciler.Trigger()
	}
}

// CreateServiceInput is the payload for adding a service.
type CreateServiceInput struct {
	ProjectID           string `json:"project_id"`
	Nome                string `json:"nome"`
	Kind                Kind   `json:"kind"`
	Target              string `json:"target"`
	ExpectedStatus      string `json:"expected_status"`
	DegradedThresholdMs int    `json:"degraded_threshold_ms"`
	TimeoutMs           int    `json:"timeout_ms"`
	IntervalSeconds     int    `json:"interval_seconds"`
	AntiFlapN           int    `json:"anti_flap_n"`
	RecoveryK           int    `json:"recovery_k"`
}

// UpdateServiceInput is a partial update of a service's configuration.
type UpdateServiceInput struct {
	Nome                *string `json:"nome"`
	Kind                *Kind   `json:"kind"`
	Target              *string `json:"target"`
	ExpectedStatus      *string `json:"expected_status"`
	DegradedThresholdMs *int    `json:"degraded_threshold_ms"`
	TimeoutMs           *int    `json:"timeout_ms"`
	IntervalSeconds     *int    `json:"interval_seconds"`
	AntiFlapN           *int    `json:"anti_flap_n"`
	RecoveryK           *int    `json:"recovery_k"`
}

// ServiceDetail bundles a service with its rolling uptime and recent checks.
type ServiceDetail struct {
	Service   *Service      `json:"service"`
	Uptime24h Uptime        `json:"uptime_24h"`
	Uptime7d  Uptime        `json:"uptime_7d"`
	Uptime30d Uptime        `json:"uptime_30d"`
	Checks    []CheckResult `json:"checks"`
}

// ListServices returns services, optionally scoped to a project ("" = all).
func (m *Manager) ListServices(ctx context.Context, projectID string) ([]*Service, error) {
	return m.repo.ListServices(ctx, projectID)
}

// GetService returns a single service.
func (m *Manager) GetService(ctx context.Context, id ServiceID) (*Service, error) {
	return m.repo.GetService(ctx, id)
}

// Detail returns a service plus rolling uptime windows and recent checks.
func (m *Manager) Detail(ctx context.Context, id ServiceID) (*ServiceDetail, error) {
	svc, err := m.repo.GetService(ctx, id)
	if err != nil {
		return nil, err
	}
	d := &ServiceDetail{Service: svc}
	if d.Uptime24h, err = m.repo.Uptime(ctx, id, 24); err != nil {
		return nil, err
	}
	if d.Uptime7d, err = m.repo.Uptime(ctx, id, 24*7); err != nil {
		return nil, err
	}
	if d.Uptime30d, err = m.repo.Uptime(ctx, id, 24*30); err != nil {
		return nil, err
	}
	if d.Checks, err = m.repo.ListChecks(ctx, id, 60); err != nil {
		return nil, err
	}
	return d, nil
}

// CreateService validates and persists a new service.
func (m *Manager) CreateService(ctx context.Context, in CreateServiceInput) (*Service, error) {
	svc := &Service{
		ID:                  ServiceID(idgen.New()),
		ProjectID:           strings.TrimSpace(in.ProjectID),
		Nome:                strings.TrimSpace(in.Nome),
		Kind:                in.Kind,
		Target:              strings.TrimSpace(in.Target),
		ExpectedStatus:      strings.TrimSpace(in.ExpectedStatus),
		DegradedThresholdMs: orDefault(in.DegradedThresholdMs, 500),
		TimeoutMs:           orDefault(in.TimeoutMs, 5000),
		IntervalSeconds:     orDefault(in.IntervalSeconds, 60),
		AntiFlapN:           orDefault(in.AntiFlapN, 3),
		RecoveryK:           orDefault(in.RecoveryK, 2),
		Enabled:             true,
		CurrentStatus:       StatusUnknown,
	}
	if svc.ExpectedStatus == "" {
		svc.ExpectedStatus = "2xx"
	}
	if err := validateService(svc); err != nil {
		return nil, err
	}
	count, err := m.repo.CountServicesByProject(ctx, svc.ProjectID)
	if err != nil {
		return nil, err
	}
	if count >= maxServicesPerProject {
		return nil, fmt.Errorf("%w: limite de %d serviços por projeto", ErrInvalid, maxServicesPerProject)
	}
	if err := m.repo.CreateService(ctx, svc); err != nil {
		return nil, err
	}
	m.notifyReconcile()
	return m.repo.GetService(ctx, svc.ID)
}

// UpdateService applies a partial config update (bumps updated_at, which makes
// the scheduler restart the service's loop with the new config).
func (m *Manager) UpdateService(ctx context.Context, id ServiceID, in UpdateServiceInput) (*Service, error) {
	svc, err := m.repo.GetService(ctx, id)
	if err != nil {
		return nil, err
	}
	if in.Nome != nil {
		svc.Nome = strings.TrimSpace(*in.Nome)
	}
	if in.Kind != nil {
		svc.Kind = *in.Kind
	}
	if in.Target != nil {
		svc.Target = strings.TrimSpace(*in.Target)
	}
	if in.ExpectedStatus != nil {
		svc.ExpectedStatus = strings.TrimSpace(*in.ExpectedStatus)
	}
	if in.DegradedThresholdMs != nil {
		svc.DegradedThresholdMs = *in.DegradedThresholdMs
	}
	if in.TimeoutMs != nil {
		svc.TimeoutMs = *in.TimeoutMs
	}
	if in.IntervalSeconds != nil {
		svc.IntervalSeconds = *in.IntervalSeconds
	}
	if in.AntiFlapN != nil {
		svc.AntiFlapN = *in.AntiFlapN
	}
	if in.RecoveryK != nil {
		svc.RecoveryK = *in.RecoveryK
	}
	if err := validateService(svc); err != nil {
		return nil, err
	}
	if err := m.repo.UpdateServiceConfig(ctx, svc); err != nil {
		return nil, err
	}
	m.notifyReconcile()
	return m.repo.GetService(ctx, id)
}

// SetEnabled toggles a service. Disabling pauses it; enabling resets it to
// unknown so derivation starts fresh.
func (m *Manager) SetEnabled(ctx context.Context, id ServiceID, enabled bool) (*Service, error) {
	svc, err := m.repo.GetService(ctx, id)
	if err != nil {
		return nil, err
	}
	svc.Enabled = enabled
	if err := m.repo.UpdateServiceConfig(ctx, svc); err != nil {
		return nil, err
	}
	status := StatusUnknown
	if !enabled {
		status = StatusPaused
	}
	if err := m.repo.SetServiceStatus(ctx, id, status, true); err != nil {
		return nil, err
	}
	m.notifyReconcile()
	return m.repo.GetService(ctx, id)
}

// DeleteService removes a service and its history.
func (m *Manager) DeleteService(ctx context.Context, id ServiceID) error {
	if err := m.repo.DeleteService(ctx, id); err != nil {
		return err
	}
	m.notifyReconcile()
	return nil
}

// ListAlerts returns recent alerts (the notification center).
func (m *Manager) ListAlerts(ctx context.Context, limit int, unreadOnly bool) ([]Alert, error) {
	return m.repo.ListAlerts(ctx, limit, unreadOnly)
}

// UnreadCount returns the number of unread alerts.
func (m *Manager) UnreadCount(ctx context.Context) (int, error) {
	return m.repo.CountUnreadAlerts(ctx)
}

// MarkAlertRead marks one alert read.
func (m *Manager) MarkAlertRead(ctx context.Context, id int64) error {
	return m.repo.MarkAlertRead(ctx, id)
}

// MarkAllAlertsRead marks every alert read.
func (m *Manager) MarkAllAlertsRead(ctx context.Context) error {
	return m.repo.MarkAllAlertsRead(ctx)
}

func validateService(s *Service) error {
	if s.ProjectID == "" {
		return fmt.Errorf("%w: project_id é obrigatório", ErrInvalid)
	}
	if strings.TrimSpace(s.Nome) == "" {
		return fmt.Errorf("%w: nome é obrigatório", ErrInvalid)
	}
	if err := validate.Var(string(s.Kind), "oneof=http tcp db_ping"); err != nil {
		return fmt.Errorf("%w: kind deve ser http, tcp ou db_ping", ErrInvalid)
	}
	if s.IntervalSeconds < 5 {
		return fmt.Errorf("%w: interval_seconds deve ser >= 5", ErrInvalid)
	}
	if s.TimeoutMs >= s.IntervalSeconds*1000 {
		return fmt.Errorf("%w: timeout_ms deve ser menor que o intervalo", ErrInvalid)
	}
	if s.DegradedThresholdMs <= 0 || s.DegradedThresholdMs >= s.TimeoutMs {
		return fmt.Errorf("%w: degraded_threshold_ms deve estar entre 1 e timeout_ms", ErrInvalid)
	}
	if s.AntiFlapN < 1 || s.RecoveryK < 1 {
		return fmt.Errorf("%w: anti_flap_n e recovery_k devem ser >= 1", ErrInvalid)
	}
	switch s.Kind {
	case KindHTTP:
		u, err := url.Parse(s.Target)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
			return fmt.Errorf("%w: target http deve ser uma URL http(s)", ErrInvalid)
		}
	case KindTCP, KindDBPing:
		if _, _, err := net.SplitHostPort(s.Target); err != nil {
			return fmt.Errorf("%w: target deve ser host:porta", ErrInvalid)
		}
	default:
		return fmt.Errorf("%w: kind inválido", ErrInvalid)
	}
	return nil
}

func orDefault(v, def int) int {
	if v == 0 {
		return def
	}
	return v
}
