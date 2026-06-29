// Package auth implements account authentication: Argon2id passwords,
// server-side opaque-token sessions, admin bootstrap, open signup with admin
// activation, login rate-limiting, and per-session CSRF tokens. The cookie holds
// only the random token; the database stores its SHA-256.
package auth

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/lumni/mirante/internal/platform/db"
	"github.com/lumni/mirante/internal/platform/id"
	"github.com/lumni/mirante/internal/platform/ratelimit"
)

// Service-level errors.
var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrRateLimited        = errors.New("too many attempts")
	ErrUnauthenticated    = errors.New("unauthenticated")
	ErrPendingApproval    = errors.New("account pending admin approval")
	ErrAccountNotActive   = errors.New("account is not active")
)

// Service ties the stores together with the login limiter.
type Service struct {
	db            *sql.DB
	users         *UserStore
	sessions      *SessionStore
	resets        *PasswordResetStore
	limiter       *ratelimit.Limiter
	resetLimiter  *ratelimit.Limiter
	signupLimiter *ratelimit.Limiter
	ttl           time.Duration

	// dummyHash is a well-formed Argon2id hash verified against when a login email
	// has no account, so the response time matches the real path and the owner's
	// e-mail can't be told apart by timing (M3). It is produced by HashPassword, so
	// it tracks the same defaultParams as every stored hash and the verify cost
	// stays equal as long as those params are shared.
	dummyHash string

	// Password-reset delivery (wired via WithMailer). A nil mailer logs the link.
	mailer       Mailer
	resetTTL     time.Duration
	resetBaseURL string
}

// NewService builds the auth service. ttl is the absolute session lifetime.
func NewService(db *sql.DB, ttl time.Duration) *Service {
	// Precompute the constant-time dummy hash once (HashPassword only fails if the
	// system CSPRNG fails, in which case the dummy verify degrades to a cheap
	// no-op — acceptable for this timing-hardening measure).
	dummy, _ := HashPassword("mirante-constant-time-placeholder")
	return &Service{
		db:            db,
		users:         NewUserStore(db),
		sessions:      NewSessionStore(db),
		resets:        NewPasswordResetStore(db),
		limiter:       ratelimit.New(5, 15*time.Minute),
		resetLimiter:  ratelimit.New(3, 15*time.Minute),
		signupLimiter: ratelimit.New(5, time.Hour),
		dummyHash:     dummy,
		ttl:           ttl,
		resetTTL:      time.Hour, // default; overridden by WithMailer
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

// Bootstrap seeds the admin from environment config if no user exists yet. It is
// idempotent. With no OWNER_EMAIL it is a no-op: the admin is then claimed by the
// first signup instead (see Signup). OWNER_EMAIL without a password/hash is a
// real misconfiguration and still errors.
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
	return s.users.Create(ctx, &User{
		ID: id.New(), Email: email, PasswordHash: hash, Role: RoleAdmin, Status: StatusActive,
	})
}

// NeedsSetup reports whether the instance has no account yet (so the UI can frame
// the first signup as creating the admin).
func (s *Service) NeedsSetup(ctx context.Context) (bool, error) {
	n, err := s.users.Count(ctx)
	return n == 0, err
}

// Signup creates a self-service account. The first account becomes the admin and
// is logged in immediately (session + token returned). Every later signup is
// created 'pending' and returns ErrPendingApproval with no session — it cannot
// log in until an admin activates it. A duplicate e-mail returns ErrEmailTaken.
func (s *Service) Signup(ctx context.Context, email, password, name, userAgent, ip string) (*Session, string, error) {
	email = strings.TrimSpace(email)
	if email == "" || password == "" {
		return nil, "", ErrInvalidCredentials
	}
	// Dedicated per-IP cap on account creation (defense-in-depth against signup
	// abuse; the takeover race itself is closed by seeding OWNER_* on deploy). The
	// IP is only meaningful behind a trusted proxy (F4); empty IP is not throttled.
	if ip != "" && !s.signupLimiter.Allow(ip) {
		return nil, "", ErrRateLimited
	}
	hash, err := HashPassword(password)
	if err != nil {
		return nil, "", err
	}
	u := &User{ID: id.New(), Email: email, Name: strings.TrimSpace(name), PasswordHash: hash}
	isFirst, err := s.users.CreateAccount(ctx, u)
	if err != nil {
		return nil, "", err
	}
	if !isFirst {
		return nil, "", ErrPendingApproval
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

	var u *User
	err := db.Retry(ctx, func() error {
		var e error
		u, e = s.users.GetByEmail(ctx, email)
		return e
	})
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			// Burn the same Argon2id cost as the real path so a missing account
			// can't be distinguished from a wrong password by response time (M3).
			_, _ = VerifyPassword(password, s.dummyHash)
			return nil, "", ErrInvalidCredentials
		}
		return nil, "", err
	}

	ok, err := VerifyPassword(password, u.PasswordHash)
	if err != nil || !ok {
		return nil, "", ErrInvalidCredentials
	}
	s.limiter.Reset(key)

	// Only an activated account may log in (revealed only after a correct password).
	if u.Status != StatusActive {
		return nil, "", ErrAccountNotActive
	}

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
	// A transient failure before the INSERT runs is safe to retry: the row was not
	// written, so the same session is re-inserted on a fresh connection.
	if err := db.Retry(ctx, func() error {
		return s.sessions.Create(ctx, sess, hashToken(token))
	}); err != nil {
		return nil, "", err
	}
	return sess, token, nil
}

// Authenticate validates a cookie token and returns the owner + session.
func (s *Service) Authenticate(ctx context.Context, token string) (*User, *Session, error) {
	if token == "" {
		return nil, nil, ErrUnauthenticated
	}
	var sess *Session
	if err := db.Retry(ctx, func() error {
		var e error
		sess, e = s.sessions.GetByToken(ctx, hashToken(token))
		return e
	}); err != nil {
		return nil, nil, ErrUnauthenticated
	}
	if sess.RevokedAt != nil || time.Now().After(sess.ExpiresAt) {
		return nil, nil, ErrUnauthenticated
	}
	var u *User
	if err := db.Retry(ctx, func() error {
		var e error
		u, e = s.users.GetByID(ctx, sess.UserID)
		return e
	}); err != nil {
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

// ListUsers returns every account (admin user management).
func (s *Service) ListUsers(ctx context.Context) ([]*User, error) {
	return s.users.List(ctx)
}

// ActivateUser marks a pending/disabled account active so it can log in.
func (s *Service) ActivateUser(ctx context.Context, userID string) error {
	return s.users.SetStatus(ctx, userID, StatusActive)
}

// DeactivateUser disables an account and revokes its live sessions.
func (s *Service) DeactivateUser(ctx context.Context, userID string) error {
	if err := s.users.SetStatus(ctx, userID, StatusDisabled); err != nil {
		return err
	}
	return s.sessions.RevokeAllForUser(ctx, userID)
}

// AdminCreateUser creates an already-active account directly (admin), bypassing
// the pending flow. role is coerced to "user" unless it is "admin".
func (s *Service) AdminCreateUser(ctx context.Context, email, password, name, role string) (*User, error) {
	email = strings.TrimSpace(email)
	if email == "" || len(password) < 8 {
		return nil, ErrInvalidCredentials
	}
	if role != RoleAdmin {
		role = RoleUser
	}
	if _, err := s.users.GetByEmail(ctx, email); err == nil {
		return nil, ErrEmailTaken
	} else if !errors.Is(err, ErrUserNotFound) {
		return nil, err
	}
	hash, err := HashPassword(password)
	if err != nil {
		return nil, err
	}
	u := &User{ID: id.New(), Email: email, Name: strings.TrimSpace(name), PasswordHash: hash, Role: role, Status: StatusActive}
	if err := s.users.Create(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

// DeleteUser removes an account and ALL of its data. libSQL runs with foreign
// keys off, so the domain rows are purged explicitly (in one transaction) rather
// than via cascade.
func (s *Service) DeleteUser(ctx context.Context, userID string) error {
	if err := s.purgeUserData(ctx, userID); err != nil {
		return err
	}
	return s.users.Delete(ctx, userID)
}

// purgeUserData deletes every row owned by userID across all domains. Monitor
// history (check_results/check_rollups) is keyed by service, so it is removed via
// the user's services before the services themselves.
func (s *Service) purgeUserData(ctx context.Context, userID string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	for _, q := range []string{
		`DELETE FROM check_results WHERE service_id IN (SELECT id FROM services WHERE user_id = ?)`,
		`DELETE FROM check_rollups WHERE service_id IN (SELECT id FROM services WHERE user_id = ?)`,
	} {
		if _, err := tx.ExecContext(ctx, q, userID); err != nil {
			return err
		}
	}
	for _, t := range []string{
		"project_links", "project_tags", "tags", "task_tags", "tasks", "projects",
		"subscriptions", "job_skills", "jobs", "applications",
		"cv_skills", "cv_experience", "cv_education", "cv_profile",
		"alerts", "events", "services", "llm_usage", "password_resets", "sessions",
	} {
		if _, err := tx.ExecContext(ctx, `DELETE FROM `+t+` WHERE user_id = ?`, userID); err != nil {
			return err
		}
	}
	return tx.Commit()
}
