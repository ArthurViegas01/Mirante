package auth

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

// Errors returned by the stores.
var (
	ErrUserNotFound    = errors.New("user not found")
	ErrSessionNotFound = errors.New("session not found")
	ErrResetNotFound   = errors.New("password reset not found")
)

const tsLayout = "2006-01-02T15:04:05.000Z"

func formatTS(t time.Time) string { return t.UTC().Format(tsLayout) }

func parseTS(s string) time.Time {
	for _, layout := range []string{tsLayout, time.RFC3339Nano, time.RFC3339} {
		if t, err := time.Parse(layout, s); err == nil {
			return t.UTC()
		}
	}
	return time.Time{}
}

// User is the single owner of the app.
type User struct {
	ID           string
	Email        string
	Name         string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Session is a server-side session; the cookie holds only the opaque token.
type Session struct {
	ID         string
	UserID     string
	CSRFToken  string
	UserAgent  string
	IP         string
	CreatedAt  time.Time
	LastUsedAt time.Time
	ExpiresAt  time.Time
	RevokedAt  *time.Time
}

// UserStore persists users.
type UserStore struct{ db *sql.DB }

// NewUserStore builds a UserStore.
func NewUserStore(db *sql.DB) *UserStore { return &UserStore{db: db} }

// Count returns the number of users (used to bootstrap exactly one owner).
func (s *UserStore) Count(ctx context.Context) (int, error) {
	var n int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&n)
	return n, err
}

// Create inserts a new user.
func (s *UserStore) Create(ctx context.Context, u *User) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO users (id, email, name, password_hash) VALUES (?, ?, ?, ?)`,
		u.ID, u.Email, nullable(u.Name), u.PasswordHash)
	return err
}

// CreateFirst inserts u only if no user exists yet, atomically. The count and
// insert run in one transaction; with the single-writer pool (MaxOpenConns=1)
// the transaction holds the only connection, so a concurrent claim blocks and
// then observes the owner. Returns ErrSignupClosed if an owner already exists.
func (s *UserStore) CreateFirst(ctx context.Context, u *User) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var n int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return ErrSignupClosed
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO users (id, email, name, password_hash) VALUES (?, ?, ?, ?)`,
		u.ID, u.Email, nullable(u.Name), u.PasswordHash); err != nil {
		return err
	}
	return tx.Commit()
}

// UpdatePassword sets a new password hash for a user and bumps updated_at.
func (s *UserStore) UpdatePassword(ctx context.Context, userID, hash string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE users SET password_hash = ?, updated_at = ? WHERE id = ?`,
		hash, formatTS(time.Now()), userID)
	return err
}

// GetByEmail looks up a user by email (case-insensitive).
func (s *UserStore) GetByEmail(ctx context.Context, email string) (*User, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, email, name, password_hash, created_at, updated_at FROM users WHERE email = ?`, email)
	return scanUser(row)
}

// GetByID looks up a user by id.
func (s *UserStore) GetByID(ctx context.Context, id string) (*User, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, email, name, password_hash, created_at, updated_at FROM users WHERE id = ?`, id)
	return scanUser(row)
}

func scanUser(row *sql.Row) (*User, error) {
	var (
		u                    User
		name                 sql.NullString
		createdAt, updatedAt string
	)
	if err := row.Scan(&u.ID, &u.Email, &name, &u.PasswordHash, &createdAt, &updatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	u.Name = name.String
	u.CreatedAt = parseTS(createdAt)
	u.UpdatedAt = parseTS(updatedAt)
	return &u, nil
}

// SessionStore persists sessions.
type SessionStore struct{ db *sql.DB }

// NewSessionStore builds a SessionStore.
func NewSessionStore(db *sql.DB) *SessionStore { return &SessionStore{db: db} }

// Create inserts a session, storing only the token hash.
func (s *SessionStore) Create(ctx context.Context, sess *Session, tokenHash string) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO sessions (id, user_id, token_hash, csrf_token, user_agent, ip, expires_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		sess.ID, sess.UserID, tokenHash, sess.CSRFToken,
		nullable(sess.UserAgent), nullable(sess.IP), formatTS(sess.ExpiresAt))
	return err
}

// GetByToken returns a session by its token hash.
func (s *SessionStore) GetByToken(ctx context.Context, tokenHash string) (*Session, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, user_id, csrf_token, user_agent, ip, created_at, last_used_at, expires_at, revoked_at
		 FROM sessions WHERE token_hash = ?`, tokenHash)

	var (
		sess                             Session
		ua, ip                           sql.NullString
		createdAt, lastUsedAt, expiresAt string
		revokedAt                        sql.NullString
	)
	if err := row.Scan(&sess.ID, &sess.UserID, &sess.CSRFToken, &ua, &ip,
		&createdAt, &lastUsedAt, &expiresAt, &revokedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}
	sess.UserAgent = ua.String
	sess.IP = ip.String
	sess.CreatedAt = parseTS(createdAt)
	sess.LastUsedAt = parseTS(lastUsedAt)
	sess.ExpiresAt = parseTS(expiresAt)
	if revokedAt.Valid {
		t := parseTS(revokedAt.String)
		sess.RevokedAt = &t
	}
	return &sess, nil
}

// Touch updates last_used_at.
func (s *SessionStore) Touch(ctx context.Context, id string, at time.Time) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE sessions SET last_used_at = ? WHERE id = ?`, formatTS(at), id)
	return err
}

// Revoke marks a session revoked by token hash.
func (s *SessionStore) Revoke(ctx context.Context, tokenHash string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE sessions SET revoked_at = ? WHERE token_hash = ? AND revoked_at IS NULL`,
		formatTS(time.Now()), tokenHash)
	return err
}

// RevokeAllForUser revokes every active session of a user. Called after a
// password reset so a leaked or lingering cookie cannot outlive the credential.
func (s *SessionStore) RevokeAllForUser(ctx context.Context, userID string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE sessions SET revoked_at = ? WHERE user_id = ? AND revoked_at IS NULL`,
		formatTS(time.Now()), userID)
	return err
}

// DeleteExpired removes sessions past their expiry (GC).
func (s *SessionStore) DeleteExpired(ctx context.Context, now time.Time) (int64, error) {
	res, err := s.db.ExecContext(ctx,
		`DELETE FROM sessions WHERE expires_at < ?`, formatTS(now))
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// PasswordReset is a single-use, time-boxed token that authorizes a password
// change. The e-mail carries the plaintext token; only its hash is stored.
type PasswordReset struct {
	ID        string
	UserID    string
	ExpiresAt time.Time
	UsedAt    *time.Time
}

// PasswordResetStore persists password-reset tokens.
type PasswordResetStore struct{ db *sql.DB }

// NewPasswordResetStore builds a PasswordResetStore.
func NewPasswordResetStore(db *sql.DB) *PasswordResetStore { return &PasswordResetStore{db: db} }

// Create inserts a reset token, storing only its hash.
func (s *PasswordResetStore) Create(ctx context.Context, pr *PasswordReset, tokenHash string) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO password_resets (id, user_id, token_hash, expires_at) VALUES (?, ?, ?, ?)`,
		pr.ID, pr.UserID, tokenHash, formatTS(pr.ExpiresAt))
	return err
}

// GetByToken returns the reset for a token hash.
func (s *PasswordResetStore) GetByToken(ctx context.Context, tokenHash string) (*PasswordReset, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, user_id, expires_at, used_at FROM password_resets WHERE token_hash = ?`, tokenHash)

	var (
		pr        PasswordReset
		expiresAt string
		usedAt    sql.NullString
	)
	if err := row.Scan(&pr.ID, &pr.UserID, &expiresAt, &usedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrResetNotFound
		}
		return nil, err
	}
	pr.ExpiresAt = parseTS(expiresAt)
	if usedAt.Valid {
		t := parseTS(usedAt.String)
		pr.UsedAt = &t
	}
	return &pr, nil
}

// MarkUsed stamps a reset as redeemed so it cannot be replayed.
func (s *PasswordResetStore) MarkUsed(ctx context.Context, id string, at time.Time) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE password_resets SET used_at = ? WHERE id = ?`, formatTS(at), id)
	return err
}

// DeleteForUser removes a user's outstanding resets (called before issuing a
// new one so only the latest link is live).
func (s *PasswordResetStore) DeleteForUser(ctx context.Context, userID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM password_resets WHERE user_id = ?`, userID)
	return err
}

// DeleteExpired removes used or expired resets (GC).
func (s *PasswordResetStore) DeleteExpired(ctx context.Context, now time.Time) (int64, error) {
	res, err := s.db.ExecContext(ctx,
		`DELETE FROM password_resets WHERE expires_at < ? OR used_at IS NOT NULL`, formatTS(now))
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func nullable(s string) any {
	if s == "" {
		return nil
	}
	return s
}
