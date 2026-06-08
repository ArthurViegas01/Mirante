// Package auth implements single-owner authentication: Argon2id passwords,
// server-side opaque-token sessions, owner bootstrap, login rate-limiting, and
// per-session CSRF tokens. The cookie holds only the random token; the database
// stores its SHA-256.
package auth

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/lumni/mirante/internal/platform/id"
	"github.com/lumni/mirante/internal/platform/ratelimit"
)

// Service-level errors.
var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrRateLimited        = errors.New("too many attempts")
	ErrUnauthenticated    = errors.New("unauthenticated")
)

// Service ties the stores together with the login limiter.
type Service struct {
	users    *UserStore
	sessions *SessionStore
	limiter  *ratelimit.Limiter
	ttl      time.Duration
}

// NewService builds the auth service. ttl is the absolute session lifetime.
func NewService(db *sql.DB, ttl time.Duration) *Service {
	return &Service{
		users:    NewUserStore(db),
		sessions: NewSessionStore(db),
		limiter:  ratelimit.New(5, 15*time.Minute),
		ttl:      ttl,
	}
}

// Bootstrap seeds the single owner if no user exists yet. It is idempotent.
func (s *Service) Bootstrap(ctx context.Context, email, password, passwordHash string) error {
	n, err := s.users.Count(ctx)
	if err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	if strings.TrimSpace(email) == "" {
		return errors.New("OWNER_EMAIL is required to bootstrap the owner")
	}
	hash := passwordHash
	if hash == "" {
		if password == "" {
			return errors.New("OWNER_PASSWORD or OWNER_PASSWORD_HASH is required to bootstrap")
		}
		hash, err = HashPassword(password)
		if err != nil {
			return err
		}
	}
	return s.users.Create(ctx, &User{ID: id.New(), Email: email, PasswordHash: hash})
}

// Login verifies credentials and creates a session, returning it plus the
// plaintext token to set in the cookie.
func (s *Service) Login(ctx context.Context, email, password, userAgent, ip string) (*Session, string, error) {
	key := strings.ToLower(strings.TrimSpace(email))
	if !s.limiter.Allow(key) {
		return nil, "", ErrRateLimited
	}

	u, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, "", ErrInvalidCredentials
		}
		return nil, "", err
	}

	ok, err := VerifyPassword(password, u.PasswordHash)
	if err != nil || !ok {
		return nil, "", ErrInvalidCredentials
	}
	s.limiter.Reset(key)

	token, err := newToken()
	if err != nil {
		return nil, "", err
	}
	csrf, err := newToken()
	if err != nil {
		return nil, "", err
	}

	now := time.Now().UTC()
	sess := &Session{
		ID:        id.New(),
		UserID:    u.ID,
		CSRFToken: csrf,
		UserAgent: userAgent,
		IP:        ip,
		ExpiresAt: now.Add(s.ttl),
	}
	if err := s.sessions.Create(ctx, sess, hashToken(token)); err != nil {
		return nil, "", err
	}
	return sess, token, nil
}

// Authenticate validates a cookie token and returns the owner + session.
func (s *Service) Authenticate(ctx context.Context, token string) (*User, *Session, error) {
	if token == "" {
		return nil, nil, ErrUnauthenticated
	}
	sess, err := s.sessions.GetByToken(ctx, hashToken(token))
	if err != nil {
		return nil, nil, ErrUnauthenticated
	}
	if sess.RevokedAt != nil || time.Now().After(sess.ExpiresAt) {
		return nil, nil, ErrUnauthenticated
	}
	u, err := s.users.GetByID(ctx, sess.UserID)
	if err != nil {
		return nil, nil, ErrUnauthenticated
	}
	_ = s.sessions.Touch(ctx, sess.ID, time.Now().UTC())
	return u, sess, nil
}

// Logout revokes the session bound to the token.
func (s *Service) Logout(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	return s.sessions.Revoke(ctx, hashToken(token))
}

// SweepExpired deletes expired sessions; intended to run periodically.
func (s *Service) SweepExpired(ctx context.Context) (int64, error) {
	return s.sessions.DeleteExpired(ctx, time.Now().UTC())
}
