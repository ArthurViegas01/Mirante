package auth

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/lumni/mirante/internal/platform/db"
	"github.com/lumni/mirante/internal/platform/id"
	"github.com/lumni/mirante/internal/platform/migrate"
)

// capturingMailer records the last message so tests can recover the reset link.
type capturingMailer struct {
	calls int
	to    string
	text  string
	html  string
}

func (m *capturingMailer) Send(_ context.Context, to, _, text, html string) error {
	m.calls++
	m.to, m.text, m.html = to, text, html
	return nil
}

var resetTokenRe = regexp.MustCompile(`token=([A-Za-z0-9_\-]+)`)

func newResetService(t *testing.T) *Service {
	t.Helper()
	ctx := context.Background()
	database, err := db.Open(ctx, ":memory:", "")
	require.NoError(t, err)
	t.Cleanup(func() { _ = database.Close() })
	require.NoError(t, migrate.Up(database.DB))
	return NewService(database.DB, time.Hour)
}

func TestPasswordResetFlow(t *testing.T) {
	ctx := context.Background()
	svc := newResetService(t)
	m := &capturingMailer{}
	svc.WithMailer(m, "https://app.example", time.Hour)

	// Claim the owner; keep the signup session token to prove revocation later.
	_, signupTok, err := svc.Signup(ctx, "owner@example.com", "old-password", "Owner", "ua", "1.2.3.4")
	require.NoError(t, err)

	// Request is case-insensitive on the e-mail and delivers exactly one message.
	require.NoError(t, svc.RequestPasswordReset(ctx, "OWNER@example.com"))
	require.Equal(t, 1, m.calls)
	require.Equal(t, "owner@example.com", m.to)

	match := resetTokenRe.FindStringSubmatch(m.text)
	require.Len(t, match, 2, "the e-mail should contain a reset link with a token")
	token := match[1]

	require.NoError(t, svc.ResetPassword(ctx, token, "new-password-123"))

	// Old password stops working; the new one logs in.
	_, _, err = svc.Login(ctx, "owner@example.com", "old-password", "ua", "ip")
	require.ErrorIs(t, err, ErrInvalidCredentials)
	_, _, err = svc.Login(ctx, "owner@example.com", "new-password-123", "ua", "ip")
	require.NoError(t, err)

	// The token cannot be replayed.
	require.ErrorIs(t, svc.ResetPassword(ctx, token, "yet-another-pass"), ErrResetTokenInvalid)

	// The session that predated the reset was revoked.
	_, _, err = svc.Authenticate(ctx, signupTok)
	require.ErrorIs(t, err, ErrUnauthenticated)
}

func TestRequestPasswordResetUnknownEmail(t *testing.T) {
	ctx := context.Background()
	svc := newResetService(t)
	m := &capturingMailer{}
	svc.WithMailer(m, "https://app.example", time.Hour)

	_, _, err := svc.Signup(ctx, "owner@example.com", "old-password", "Owner", "ua", "ip")
	require.NoError(t, err)

	// No account for this address → success response, but nothing is sent.
	require.NoError(t, svc.RequestPasswordReset(ctx, "stranger@example.com"))
	require.Equal(t, 0, m.calls)
}

func TestRequestPasswordResetWithoutMailer(t *testing.T) {
	ctx := context.Background()
	svc := newResetService(t)
	svc.WithMailer(nil, "https://app.example", time.Hour) // dev path: link is logged

	_, _, err := svc.Signup(ctx, "owner@example.com", "old-password", "Owner", "ua", "ip")
	require.NoError(t, err)

	// A nil mailer must not error or panic; the token is still persisted.
	require.NoError(t, svc.RequestPasswordReset(ctx, "owner@example.com"))
}

func TestResetPasswordRejectsExpiredToken(t *testing.T) {
	ctx := context.Background()
	svc := newResetService(t)

	_, _, err := svc.Signup(ctx, "owner@example.com", "old-password", "Owner", "ua", "ip")
	require.NoError(t, err)
	u, err := svc.users.GetByEmail(ctx, "owner@example.com")
	require.NoError(t, err)

	token, err := newToken()
	require.NoError(t, err)
	pr := &PasswordReset{ID: id.New(), UserID: u.ID, ExpiresAt: time.Now().UTC().Add(-time.Minute)}
	require.NoError(t, svc.resets.Create(ctx, pr, hashToken(token)))

	require.ErrorIs(t, svc.ResetPassword(ctx, token, "new-password-123"), ErrResetTokenInvalid)
}

func TestResetPasswordRejectsShortPassword(t *testing.T) {
	svc := newResetService(t)
	require.ErrorIs(t, svc.ResetPassword(context.Background(), "any-token", "short"), ErrInvalidCredentials)
}
