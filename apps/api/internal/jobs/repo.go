package jobs

import (
	"context"
	"errors"
)

// ErrNotFound is returned when a job does not exist.
var ErrNotFound = errors.New("job not found")

// Repository persists jobs and their extracted skills.
type Repository interface {
	Create(ctx context.Context, j *Job) error
	Get(ctx context.Context, id ID) (*Job, error)
	List(ctx context.Context) ([]*Job, error)
	Update(ctx context.Context, j *Job) error
	Delete(ctx context.Context, id ID) error

	SetSkills(ctx context.Context, jobID ID, skills []string) error
}
