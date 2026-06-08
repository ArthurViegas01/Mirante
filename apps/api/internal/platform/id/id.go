// Package id generates time-ordered identifiers (UUIDv7) used as primary keys.
// UUIDv7 is monotonic-ish, giving good B-tree locality on inserts while staying
// opaque in URLs.
package id

import "github.com/google/uuid"

// New returns a new UUIDv7 as a canonical string.
func New() string {
	return uuid.Must(uuid.NewV7()).String()
}
