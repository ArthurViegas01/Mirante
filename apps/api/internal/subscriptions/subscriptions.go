// Package subscriptions owns recurring costs attached to a project (a paid stack
// piece, a domain, a SaaS, a paid API). A subscription may soft-link a monitor
// service via ServiceID to show cost beside live status, but this domain does not
// import monitor (ADR-0001): ServiceID is a plain string. Money is an integer in
// the currency's minor unit; totals are summed per currency, never converted.
package subscriptions

import "time"

// ID is a subscription identifier.
type ID string

// Currency is the billing currency (no conversion between them).
type Currency string

const (
	MoedaBRL Currency = "BRL"
	MoedaUSD Currency = "USD"
)

// Cycle is the billing cadence.
type Cycle string

const (
	CicloMensal Cycle = "mensal"
	CicloAnual  Cycle = "anual"
)

// Subscription is a recurring cost line for a project.
type Subscription struct {
	ID         ID        `json:"id"`
	ProjectID  string    `json:"project_id"`
	ServiceID  string    `json:"service_id"` // optional monitor service ("" = none); no FK
	Nome       string    `json:"nome"`
	Provider   string    `json:"provider"`
	ValorCents int       `json:"valor_cents"` // amount in the currency's minor unit
	Moeda      Currency  `json:"moeda"`
	Ciclo      Cycle     `json:"ciclo"`
	Ativo      bool      `json:"ativo"`
	Notas      string    `json:"notas"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
