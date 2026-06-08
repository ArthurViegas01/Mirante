// Package tasks owns activities the user tracks — each optionally linked to a
// project and, from F3, to a job. It reuses the shared tag vocabulary.
package tasks

import "time"

// ID is a task identifier.
type ID string

// Status is a task's position on the board.
type Status string

const (
	StatusAFazer  Status = "a_fazer"
	StatusFazendo Status = "fazendo"
	StatusFeito   Status = "feito"
)

// Prioridade is a task's priority.
type Prioridade string

const (
	PrioBaixa Prioridade = "baixa"
	PrioMedia Prioridade = "media"
	PrioAlta  Prioridade = "alta"
)

// Task is an activity. ProjectID is a nullable FK→projects; JobID is nullable
// with no FK yet (the constraint FK→jobs lands in F3). Prazo is a calendar date
// (YYYY-MM-DD), nullable — stored as text to stay timezone-stable, unlike the
// system timestamps CreatedAt/UpdatedAt.
type Task struct {
	ID         ID         `json:"id"`
	Titulo     string     `json:"titulo"`
	Status     Status     `json:"status"`
	Prioridade Prioridade `json:"prioridade"`
	Prazo      *string    `json:"prazo"`
	ProjectID  *string    `json:"project_id"`
	JobID      *string    `json:"job_id"`
	Tags       []string   `json:"tags"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}
