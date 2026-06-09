// Package jobs owns the Vagas domain: job postings the owner tracks, with the
// required skills extracted from the posting text via the skills kernel and an
// optional LLM enrichment (empresa/senioridade/modelo/resumo). It consumes the
// shared kernel (internal/skills) and the LLM gateway (internal/llm); it does not
// import other domains (ADR-0001). Aderência scoring against a CV arrives with
// the cv domain.
package jobs

import "time"

// ID is a job identifier.
type ID string

// Modelo is the work arrangement of a posting.
type Modelo string

const (
	ModeloRemoto     Modelo = "remoto"
	ModeloHibrido    Modelo = "hibrido"
	ModeloPresencial Modelo = "presencial"
	ModeloIndefinido Modelo = "indefinido"
)

// Job is a tracked posting.
type Job struct {
	ID          ID        `json:"id"`
	Titulo      string    `json:"titulo"`
	Empresa     string    `json:"empresa"`
	Descricao   string    `json:"descricao"`
	URL         string    `json:"url"`
	Localizacao string    `json:"localizacao"`
	Modelo      Modelo    `json:"modelo"`
	Senioridade string    `json:"senioridade"`
	Resumo      string    `json:"resumo"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Skills      []string  `json:"skills"` // canonical skills required, extracted from Descricao
}
