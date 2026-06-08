package httpserver

import (
	"context"

	"github.com/lumni/mirante/internal/platform/auth"
)

type ctxKey int

const (
	ctxRequestID ctxKey = iota
	ctxUser
	ctxSession
)

func withRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ctxRequestID, id)
}

// RequestIDFrom returns the request id bound to the context, if any.
func RequestIDFrom(ctx context.Context) string {
	id, _ := ctx.Value(ctxRequestID).(string)
	return id
}

func withAuth(ctx context.Context, u *auth.User, s *auth.Session) context.Context {
	ctx = context.WithValue(ctx, ctxUser, u)
	return context.WithValue(ctx, ctxSession, s)
}

// UserFrom returns the authenticated owner bound to the context.
func UserFrom(ctx context.Context) (*auth.User, bool) {
	u, ok := ctx.Value(ctxUser).(*auth.User)
	return u, ok
}

// SessionFrom returns the session bound to the context.
func SessionFrom(ctx context.Context) (*auth.Session, bool) {
	s, ok := ctx.Value(ctxSession).(*auth.Session)
	return s, ok
}
