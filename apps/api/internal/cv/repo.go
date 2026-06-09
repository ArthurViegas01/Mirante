package cv

import "context"

// Repository persists the singleton master CV. GetProfile returns an empty CV
// (not an error) when nothing has been saved yet. SaveCV fully replaces the CV
// (identity + skills + experiences + education) atomically.
type Repository interface {
	GetProfile(ctx context.Context) (*Profile, error)
	SaveCV(ctx context.Context, p *Profile) error
}
