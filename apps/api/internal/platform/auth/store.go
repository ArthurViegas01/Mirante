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
	ErrEmailTaken      = errors.New("email already in use")
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

// Role and account status values.
const (
	RoleAdmin = "admin"
	RoleUser  = "user"

	StatusPending  = "pending"
	StatusActive   = "active"
	StatusDisabled = "disabled"
)

// User is an account. The first account (env-bootstrapped or first signup) is the
// admin; the rest sign up 'pending' and need activation before they can log in.
type User struct {
	ID           string
	Email        string
	Name         string
	PasswordHash string
	Role         string
	Status       string
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

// Create inserts a new user with the role and status set on u (used by the
// env bootstrap and by an admin creating an account directly).
func (s *UserStore) Create(ctx context.Context, u *User) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO users (id, email, name, password_hash, role, status) VALUES (?, ?, ?, ?, ?, ?)`,
		u.ID, u.Email, nullable(u.Name), u.PasswordHash, u.Role, u.Status)
	return err
}

// CreateAccount inserts a self-service signup atomically, deciding the role and
// status from whether an account already exists: the very first account is the
// admin and active; the rest start as a pending user. It sets u.Role/u.Status and
// reports whether this was the first account. Returns ErrEmailTaken on a
// duplicate e-mail. With the single-writer pool the count+insert can't race.
func (s *UserStore) CreateAccount(ctx context.Context, u *User) (isFirst bool, err error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer func() { _ = tx.Rollback() }()

	var n int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&n); err != nil {
		return false, err
	}
	var exists int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE email = ?`, u.Email).Scan(&exists); err != nil {
		return false, err
	}
	if exists > 0 {
		return false, ErrEmailTaken
	}

	isFirst = n == 0
	if isFirst {
		u.Role, u.Status = RoleAdmin, StatusActive
	} else {
		u.Role, u.Status = RoleUser, StatusPending
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO users (id, email, name, password_hash, role, status) VALUES (?, ?, ?, ?, ?, ?)`,
		u.ID, u.Email, nullable(u.Name), u.PasswordHash, u.Role, u.Status); err != nil {
		return false, err
	}
	return isFirst, tx.Commit()
}

// List returns all users, newest first (admin user management).
func (s *UserStore) List(ctx context.Context) ([]*User, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, email, name, password_hash, role, status, created_at, updated_at
		 FROM users ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := []*User{}
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

// SetStatus updates a user's account status (admin activate/deactivate).
func (s *UserStore) SetStatus(ctx context.Context, userID, status string) error {
	res, err := s.db.ExecContext(ctx,
		`UPDATE users SET status = ?, updated_at = ? WHERE id = ?`, status, formatTS(time.Now()), userID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrUserNotFound
	}
	return nil
}

// Delete removes a user row (the caller purges the user's domain data first).
func (s *UserStore) Delete(ctx context.Context, userID string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM users WHERE id = ?`, userID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrUserNotFound
	}
	return nil
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
		`SELECT id, email, name, password_hash, role, status, created_at, updated_at FROM users WHERE email = ?`, email)
	return scanUser(row)
}

// GetByID looks up a user by id.
func (s *UserStore) GetByID(ctx context.Context, id string) (*User, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, email, name, password_hash, role, status, created_at, updated_at FROM users WHERE id = ?`, id)
	return scanUser(row)
}

func scanUser(row interface{ Scan(dest ...any) error }) (*User, error) {
	var (
		u                    User
		name                 sql.NullString
		createdAt, updatedAt string
	)
	if err := row.Scan(&u.ID, &u.Email, &name, &u.PasswordHash, &u.Role, &u.Status, &createdAt, &updatedAt); err != nil {
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
