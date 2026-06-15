// Package tenant carries the authenticated user's id in the request context so
// data-access code can scope every query to its owner. The auth middleware sets
// it right after authenticating; domain repositories read it to isolate rows per
// user. Keeping it in a tiny platform package (not httpserver) lets domains read
// the owner without importing the web layer.
package tenant

import "context"

type ctxKey struct{}

// WithUserID returns a context carrying the owner's user id.
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, ctxKey{}, userID)
}

// UserID returns the owner's user id from the context. ok is false when no user
// is set or the id is empty — callers scoping a query should treat that as "no
// rows" rather than "all rows".
func UserID(ctx context.Context) (id string, ok bool) {
	id, _ = ctx.Value(ctxKey{}).(string)
	return id, id != ""
}
