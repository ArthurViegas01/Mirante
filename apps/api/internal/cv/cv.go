// Package cv owns the owner's master CV. Today it exposes the singleton profile
// (identity + headline) consumed across the career-search area — e.g. the target
// role shown on the Vagas header. The full master CV (experiences, education,
// skills, per-vaga adaptation, PDF/DOCX export, aderência) grows from here. Per
// ADR-0001 it does not import other domains.
package cv

import "time"

// Profile is the owner's master profile (a singleton in this single-user app).
type Profile struct {
	Nome       string    `json:"nome"`
	Titulo     string    `json:"titulo"`      // current headline / profession
	TituloAlvo string    `json:"titulo_alvo"` // role being aimed for
	Resumo     string    `json:"resumo"`
	UpdatedAt  time.Time `json:"updated_at"`
}
