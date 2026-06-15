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
	ErrSignupClosed       = errors.New("registration closed: owner already exists")
)

// Service ties the stores together with the login limiter.
type Service struct {
	users        *UserStore
	sessions     *SessionStore
	resets       *PasswordResetStore
	limiter      *ratelimit.Limiter
	resetLimiter *ratelimit.Limiter
	ttl          time.Duration

	// Password-reset delivery (wired via WithMailer). A nil mailer logs the link.
	mailer       Mailer
	resetTTL     time.Duration
	resetBaseURL string
}

// NewService builds the auth service. ttl is the absolute session lifetime.
func NewService(db *sql.DB, ttl time.Duration) *Service {
	return &Service{
		users:        NewUserStore(db),
		sessions:     NewSessionStore(db),
		resets:       NewPasswordResetStore(db),
		limiter:      ratelimit.New(5, 15*time.Minute),
		resetLimiter: ratelimit.New(3, 15*time.Minute),
		ttl:          ttl,
		resetTTL:     time.Hour, // default; overridden by WithMailer
	}
}

// WithMailer wires optional e-mail delivery for password resets and returns the
// service for chaining. baseURL is the web origin used to build the reset link;
// resetTTL is how long a link stays valid (<=0 keeps the default). A nil mailer
// is valid: the reset link is logged instead of e-mailed (dev).
func (s *Service) WithMailer(m Mailer, baseURL string, resetTTL time.Duration) *Service {
	s.mailer = m
	s.resetBaseURL = strings.TrimRight(baseURL, "/")
	if resetTTL > 0 {
		s.resetTTL = resetTTL
	}
	return s
}

// Bootstrap seeds the single owner from environment config if no user exists yet.
// It is idempotent. With no OWNER_EMAIL it is a no-op: the owner is then claimed
// through the first-run signup flow instead (see Signup). OWNER_EMAIL without a
// password/hash is a real misconfiguration and still errors.
func (s *Service) Bootstrap(ctx context.Context, email, password, passwordHash string) error {
	n, err := s.users.Count(ctx)
	if err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	if strings.TrimSpace(email) == "" {
		return nil
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

// NeedsSetup reports whether the instance has no owner yet (so the UI should
// route to the first-run signup instead of login).
func (s *Service) NeedsSetup(ctx context.Context) (bool, error) {
	n, err := s.users.Count(ctx)
	return n == 0, err
}

// Signup claims the instance: it creates the single owner (only if none exists
// yet) and immediately opens a session, returning it plus the cookie token. A
// second attempt once the owner exists returns ErrSignupClosed.
func (s *Service) Signup(ctx context.Context, email, password, name, userAgent, ip string) (*Session, string, error) {
	email = strings.TrimSpace(email)
	if email == "" || password == "" {
		return nil, "", ErrInvalidCredentials
	}
	hash, err := HashPassword(password)
	if err != nil {
		return nil, "", err
	}
	u := &User{ID: id.New(), Email: email, Name: strings.TrimSpace(name), PasswordHash: hash}
	if err := s.users.CreateFirst(ctx, u); err != nil {
		return nil, "", err
	}
	return s.createSession(ctx, u.ID, userAgent, ip)
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

	return s.createSession(ctx, u.ID, userAgent, ip)
}

// createSession mints a session (opaque token + CSRF token) for a user and
// persists it, returning the session and the plaintext cookie token.
func (s *Service) createSession(ctx context.Context, userID, userAgent, ip string) (*Session, string, error) {
	token, err := newToken()
	if err != nil {
		return nil, "", err
	}
	csrf, err := newToken()
	if err != nil {
		return nil, "", err
	}
	sess := &Session{
		ID:        id.New(),
		UserID:    userID,
		CSRFToken: csrf,
		UserAgent: userAgent,
		IP:        ip,
		ExpiresAt: time.Now().UTC().Add(s.ttl),
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

// SweepExpired deletes expired sessions and spent/expired reset tokens; intended
// to run periodically. It returns the number of sessions removed.
func (s *Service) SweepExpired(ctx context.Context) (int64, error) {
	now := time.Now().UTC()
	n, err := s.sessions.DeleteExpired(ctx, now)
	if err != nil {
		return n, err
	}
	if _, err := s.resets.DeleteExpired(ctx, now); err != nil {
		return n, err
	}
	return n, nil
}
