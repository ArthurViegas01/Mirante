// Package cv owns the owner's master CV. Today it exposes the singleton profile
// (identity + headline) consumed across the career-search area — e.g. the target
// role shown on the Vagas header. The full master CV (experiences, education,
// skills, per-vaga adaptation, PDF/DOCX export, aderência) grows from here. Per
// ADR-0001 it does not import other domains.
package cv

import "time"

// Profile is the owner's master CV (a singleton in this single-user app):
// identity + skills + experiences + education.
type Profile struct {
	Nome        string       `json:"nome"`
	Titulo      string       `json:"titulo"`      // current headline / profession
	TituloAlvo  string       `json:"titulo_alvo"` // role being aimed for
	Contato     string       `json:"contato"`     // email · phone · location · links
	Resumo      string       `json:"resumo"`
	Skills      []string     `json:"skills"` // master skills (canonical when recognized)
	Experiences []Experience `json:"experiences"`
	Education   []Education  `json:"education"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// Experience is one job entry in the master CV.
type Experience struct {
	ID        string `json:"id"`
	Empresa   string `json:"empresa"`
	Cargo     string `json:"cargo"`
	Inicio    string `json:"inicio"` // free text ("YYYY-MM", "2023"…)
	Fim       string `json:"fim"`    // "" = atual
	Descricao string `json:"descricao"`
}

// Education is one study entry in the master CV.
type Education struct {
	ID          string `json:"id"`
	Instituicao string `json:"instituicao"`
	Curso       string `json:"curso"`
	Inicio      string `json:"inicio"`
	Fim         string `json:"fim"`
}
