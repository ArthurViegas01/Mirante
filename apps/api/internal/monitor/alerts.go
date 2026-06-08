package monitor

import (
	"encoding/json"

	idb "github.com/lumni/mirante/internal/platform/db"
)

// buildAlert constructs the in-app alert for a transition. `from` is the prior
// status; the title uses only the service name (never the target/credentials).
func buildAlert(svc *Service, from, to Status, reason string) Alert {
	return Alert{
		ServiceID:  svc.ID,
		ProjectID:  svc.ProjectID,
		Severity:   severityFor(to),
		Title:      alertTitle(svc.Nome, to),
		Body:       reason,
		FromStatus: from,
		ToStatus:   to,
	}
}

func alertTitle(nome string, to Status) string {
	switch to {
	case StatusUp:
		return nome + " está no ar"
	case StatusDegraded:
		return nome + " está degradado"
	case StatusDown:
		return nome + " está fora do ar"
	default:
		return nome + " mudou de estado"
	}
}

// transitionEvent is the JSON payload streamed over SSE on a transition.
type transitionEvent struct {
	ServiceID ServiceID `json:"service_id"`
	ProjectID string    `json:"project_id"`
	Nome      string    `json:"nome"`
	From      Status    `json:"from"`
	To        Status    `json:"to"`
	LatencyMs int       `json:"latency_ms"`
	AlertID   int64     `json:"alert_id"`
	Severity  string    `json:"severity"`
	At        string    `json:"at"`
}

func eventData(a Alert, nome string, latencyMs int) ([]byte, error) {
	return json.Marshal(transitionEvent{
		ServiceID: a.ServiceID,
		ProjectID: a.ProjectID,
		Nome:      nome,
		From:      a.FromStatus,
		To:        a.ToStatus,
		LatencyMs: latencyMs,
		AlertID:   a.ID,
		Severity:  a.Severity,
		At:        idb.FormatTime(a.CreatedAt),
	})
}
