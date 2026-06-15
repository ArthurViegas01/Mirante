package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/lumni/mirante/internal/platform/id"
)

// ErrResetTokenInvalid is returned when a reset token is unknown, already used,
// or expired. One error covers all three so the API never reveals which.
var ErrResetTokenInvalid = errors.New("invalid or expired reset token")

// Mailer delivers transactional e-mail. It is satisfied by
// platform/mailer.SMTP. A nil Mailer makes the reset flow log the link instead
// of sending it (handy in dev, where no SMTP is configured).
type Mailer interface {
	Send(ctx context.Context, to, subject, text, html string) error
}

// htmlEscaper neutralises the few markup-significant runes in interpolated
// e-mail values (the owner's name, the link). Avoids pulling in html/template
// for two substitutions.
var htmlEscaper = strings.NewReplacer(
	"&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;", "'", "&#39;")

// RequestPasswordReset issues a single-use reset link for the account with the
// given e-mail and delivers it. To avoid revealing whether an address has an
// account, it returns nil whether or not a user was found (and silently no-ops
// when the per-address rate limit is hit). Only genuine internal failures error.
func (s *Service) RequestPasswordReset(ctx context.Context, email string) error {
	key := strings.ToLower(strings.TrimSpace(email))
	if key == "" || !s.resetLimiter.Allow(key) {
		return nil
	}

	u, err := s.users.GetByEmail(ctx, key)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil // no account enumeration
		}
		return err
	}

	token, err := newToken()
	if err != nil {
		return err
	}
	// Only the most recent link stays live.
	if err := s.resets.DeleteForUser(ctx, u.ID); err != nil {
		return err
	}
	pr := &PasswordReset{ID: id.New(), UserID: u.ID, ExpiresAt: time.Now().UTC().Add(s.resetTTL)}
	if err := s.resets.Create(ctx, pr, hashToken(token)); err != nil {
		return err
	}

	link := s.resetBaseURL + "/reset-password?token=" + token
	if s.mailer == nil {
		slog.Warn("password reset requested but no mailer configured — use the link below",
			"email", u.Email, "reset_url", link, "expires_in", s.resetTTL.String())
		return nil
	}

	subject, text, html := resetEmail(u.Name, link, s.resetTTL)
	if err := s.mailer.Send(ctx, u.Email, subject, text, html); err != nil {
		// Don't surface delivery failures to the caller (no enumeration, no leak
		// of provider state). Log it for the operator instead.
		slog.Error("failed to send password reset e-mail", "email", u.Email, "err", err)
	}
	return nil
}

// ResetPassword consumes a reset token and sets a new password, then revokes
// every session so old cookies cannot outlive the changed credential.
func (s *Service) ResetPassword(ctx context.Context, token, newPassword string) error {
	if len(newPassword) < 8 {
		return ErrInvalidCredentials
	}
	if token == "" {
		return ErrResetTokenInvalid
	}

	pr, err := s.resets.GetByToken(ctx, hashToken(token))
	if err != nil {
		if errors.Is(err, ErrResetNotFound) {
			return ErrResetTokenInvalid
		}
		return err
	}
	if pr.UsedAt != nil || time.Now().After(pr.ExpiresAt) {
		return ErrResetTokenInvalid
	}

	hash, err := HashPassword(newPassword)
	if err != nil {
		return err
	}
	if err := s.users.UpdatePassword(ctx, pr.UserID, hash); err != nil {
		return err
	}
	if err := s.resets.MarkUsed(ctx, pr.ID, time.Now().UTC()); err != nil {
		return err
	}
	// Best-effort: a failure to revoke doesn't undo the password change.
	_ = s.sessions.RevokeAllForUser(ctx, pr.UserID)
	return nil
}

// resetEmail renders the password-reset message (subject, plain text, HTML).
func resetEmail(name, link string, ttl time.Duration) (subject, text, html string) {
	greeting := "Olá"
	if n := strings.TrimSpace(name); n != "" {
		greeting = "Olá, " + n
	}
	mins := int(ttl.Minutes())

	subject = "Redefinição de senha — Mirante"
	text = fmt.Sprintf(`%s,

Recebemos um pedido para redefinir a senha da sua conta no Mirante.
Abra o link abaixo para criar uma nova senha (expira em %d minutos):

%s

Se você não fez esse pedido, ignore este e-mail — sua senha continua a mesma.

— Mirante
`, greeting, mins, link)

	html = fmt.Sprintf(`<!doctype html>
<html lang="pt-BR">
  <body style="margin:0;background:#0f1115;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,Helvetica,Arial,sans-serif;">
    <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" style="padding:32px 16px;">
      <tr><td align="center">
        <table role="presentation" cellpadding="0" cellspacing="0" style="max-width:440px;width:100%%;background:#171a21;border:1px solid #262b36;border-radius:16px;padding:32px;">
          <tr><td style="color:#e7e9ee;font-size:18px;font-weight:600;padding-bottom:8px;">Redefinição de senha</td></tr>
          <tr><td style="color:#aab1bf;font-size:14px;line-height:1.6;padding-bottom:24px;">%s, recebemos um pedido para redefinir a senha da sua conta no Mirante. O link expira em %d minutos.</td></tr>
          <tr><td style="padding-bottom:24px;">
            <a href="%s" style="display:inline-block;background:#5eead4;color:#0f1115;text-decoration:none;font-weight:600;font-size:14px;padding:12px 20px;border-radius:10px;">Criar nova senha</a>
          </td></tr>
          <tr><td style="color:#6b7280;font-size:12px;line-height:1.6;">Se você não fez esse pedido, ignore este e-mail — sua senha continua a mesma.</td></tr>
        </table>
      </td></tr>
    </table>
  </body>
</html>`, htmlEscaper.Replace(greeting), mins, htmlEscaper.Replace(link))

	return subject, text, html
}
