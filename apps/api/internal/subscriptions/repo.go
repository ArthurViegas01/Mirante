package subscriptions

import (
	"context"
	"errors"
)

// ErrNotFound is returned when a subscription does not exist.
var ErrNotFound = errors.New("subscription not found")

// ListFilter narrows a subscription listing. Empty fields are ignored.
type ListFilter struct {
	ProjectID string // optional exact project
}

// Repository persists subscriptions.
type Repository interface {
	Create(ctx context.Context, s *Subscription) error
	Get(ctx context.Context, id ID) (*Subscription, error)
	List(ctx context.Context, f ListFilter) ([]*Subscription, error)
	Update(ctx context.Context, s *Subscription) error
	Delete(ctx context.Context, id ID) error
}
