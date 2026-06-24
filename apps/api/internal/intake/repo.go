package intake

import (
	"context"
	"errors"
)

// ErrNotFound is returned when a staged item does not exist for this user.
var ErrNotFound = errors.New("intake item not found")

// ListFilter narrows a staging listing. The zero value lists everything.
type ListFilter struct {
	Estado   Estado // "" = any lifecycle state
	MinScore int    // 0 = no floor; >0 keeps only the shortlist (score >= MinScore)
}

// Repository persists staged intake items, scoped per user.
type Repository interface {
	// Upsert inserts the item when (user, fonte, fonte_id) is new, reporting
	// inserted=true. An already-staged project is left untouched (inserted=false)
	// so dedup across recurring digests never clobbers triage state.
	Upsert(ctx context.Context, it *Item) (inserted bool, err error)
	List(ctx context.Context, f ListFilter) ([]*Item, error)
	Get(ctx context.Context, id ID) (*Item, error)
	SetEstado(ctx context.Context, id ID, estado Estado) error
}
