package tasks

import (
	"context"
	"errors"
)

// ErrNotFound is returned when a task does not exist.
var ErrNotFound = errors.New("task not found")

// ListFilter narrows a task listing. Empty fields are ignored.
type ListFilter struct {
	Status     string // optional exact status
	ProjectID  string // optional exact project_id
	Prioridade string // optional exact priority
}

// Repository persists tasks and their tags.
type Repository interface {
	Create(ctx context.Context, t *Task) error
	Get(ctx context.Context, id ID) (*Task, error)
	List(ctx context.Context, f ListFilter) ([]*Task, error)
	Update(ctx context.Context, t *Task) error
	Delete(ctx context.Context, id ID) error

	SetTags(ctx context.Context, taskID ID, names []string) error
}
