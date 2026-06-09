package applications

import (
	"context"
	"errors"
)

// ErrNotFound is returned when an application does not exist.
var ErrNotFound = errors.New("application not found")

// ListFilter narrows a listing. Empty fields are ignored.
type ListFilter struct {
	Status string // optional exact status
}

// Repository persists applications.
type Repository interface {
	Create(ctx context.Context, a *Application) error
	Get(ctx context.Context, id ID) (*Application, error)
	List(ctx context.Context, f ListFilter) ([]*Application, error)
	Update(ctx context.Context, a *Application) error
	Delete(ctx context.Context, id ID) error
}
