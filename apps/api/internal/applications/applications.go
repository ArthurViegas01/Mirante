// Package applications owns the candidaturas CRM: the owner's job applications and
// their pipeline. An application references a vaga by JobID (a plain string — this
// domain does not import jobs, per ADR-0001) and snapshots the vaga's titulo/empresa
// so it survives the vaga being deleted.
package applications

import "time"

// ID is an application identifier.
type ID string

// Status is the pipeline stage of a candidatura.
type Status string

const (
	StatusInteresse  Status = "interesse"
	StatusAplicado   Status = "aplicado"
	StatusEntrevista Status = "entrevista"
	StatusOferta     Status = "oferta"
	StatusAceito     Status = "aceito"
	StatusRejeitado  Status = "rejeitado"
)

// Application is one tracked candidatura.
type Application struct {
	ID          ID        `json:"id"`
	JobID       string    `json:"job_id"`  // optional soft link to jobs.ID
	Titulo      string    `json:"titulo"`  // snapshot of the vaga title
	Empresa     string    `json:"empresa"` // snapshot
	Status      Status    `json:"status"`
	Notas       string    `json:"notas"`
	ProximaAcao string    `json:"proxima_acao"` // next follow-up note
	DataAcao    string    `json:"data_acao"`    // next follow-up date, "YYYY-MM-DD"
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
