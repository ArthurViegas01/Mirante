package projects

import (
	"context"
	"errors"
)

// ErrNotFound is returned when a project does not exist.
var ErrNotFound = errors.New("project not found")

// ListFilter narrows a project listing.
type ListFilter struct {
	Status string // optional exact status
}

// Repository persists projects, their links and tags.
type Repository interface {
	Create(ctx context.Context, p *Project) error
	Get(ctx context.Context, id ID) (*Project, error)
	List(ctx context.Context, f ListFilter) ([]*Project, error)
	Update(ctx context.Context, p *Project) error
	Delete(ctx context.Context, id ID) error

	AddLink(ctx context.Context, l *Link) error
	RemoveLink(ctx context.Context, projectID, linkID ID) error

	SetTags(ctx context.Context, projectID ID, names []string) error
}
