package cv

import "context"

// Repository persists the singleton profile. GetProfile returns an empty Profile
// (not an error) when none has been saved yet.
type Repository interface {
	GetProfile(ctx context.Context) (*Profile, error)
	SaveProfile(ctx context.Context, p *Profile) error
}
