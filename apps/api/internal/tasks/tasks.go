// Package tasks owns work items — kanban-style tasks that can hang off a project
// (and, from F3, a job). It reuses the shared tag vocabulary introduced by the
// projects migration. Per ADR-0001, this domain does not import projects: the
// project and job references are plain string IDs.
package tasks

import "time"

// ID is a task identifier.
type ID string

// Status is the kanban column a task sits in.
type Status string

const (
	StatusAFazer  Status = "a_fazer"
	StatusFazendo Status = "fazendo"
	StatusFeito   Status = "feito"
)

// Priority ranks a task's urgency.
type Priority string

const (
	PrioridadeBaixa Priority = "baixa"
	PrioridadeMedia Priority = "media"
	PrioridadeAlta  Priority = "alta"
)

// Task is a single work item.
type Task struct {
	ID         ID        `json:"id"`
	Titulo     string    `json:"titulo"`
	Descricao  string    `json:"descricao"`
	Status     Status    `json:"status"`
	Prioridade Priority  `json:"prioridade"`
	Prazo      string    `json:"prazo"`      // optional due date, "YYYY-MM-DD" ("" = none)
	ProjectID  string    `json:"project_id"` // optional projects.ID ("" = unlinked)
	JobID      string    `json:"job_id"`     // optional; the FK to jobs arrives in F3
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	Tags       []string  `json:"tags"`
}
