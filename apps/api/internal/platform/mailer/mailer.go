// Package mailer sends transactional e-mail over SMTP. It supports STARTTLS
// (typically port 587) and implicit TLS (port 465), with optional PLAIN auth —
// enough to reach the common transactional providers (SendGrid, Mailgun,
// Postmark, Gmail SMTP) and dev catchers (Mailpit/MailHog, no auth).
package mailer

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"mime"
	"mime/quotedprintable"
	"net"
	"net/mail"
	"net/smtp"
	"strconv"
	"strings"
	"time"
)

// dialTimeout bounds each phase of the SMTP conversation.
const dialTimeout = 10 * time.Second

// Config holds SMTP settings. From is the message sender, e.g.
// "Mirante <no-reply@example.com>" or a bare address.
type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

// SMTP is an SMTP-backed mailer. The zero value is not usable; build with NewSMTP.
type SMTP struct {
	cfg      Config
	addr     string
	fromAddr string // bare address parsed out of cfg.From for the envelope
}

// NewSMTP validates the config and returns a mailer.
func NewSMTP(cfg Config) (*SMTP, error) {
	if strings.TrimSpace(cfg.Host) == "" {
		return nil, errors.New("mailer: SMTP_HOST is required")
	}
	if cfg.Port == 0 {
		cfg.Port = 587
	}
	from, err := mail.ParseAddress(cfg.From)
	if err != nil {
		return nil, fmt.Errorf("mailer: invalid SMTP_FROM %q: %w", cfg.From, err)
	}
	return &SMTP{
		cfg:      cfg,
		addr:     net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port)),
		fromAddr: from.Address,
	}, nil
}

// Send delivers a text+HTML message to a single recipient. It honours ctx for
// cancellation; the underlying connection also carries its own timeouts.
func (s *SMTP) Send(ctx context.Context, to, subject, text, html string) error {
	rcpt, err := mail.ParseAddress(to)
	if err != nil {
		return fmt.Errorf("mailer: invalid recipient %q: %w", to, err)
	}
	msg := s.build(rcpt.Address, subject, text, html)

	done := make(chan error, 1)
	go func() { done <- s.deliver(rcpt.Address, msg) }()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		return err
	}
}

// deliver runs the SMTP conversation, choosing implicit TLS for port 465 and
// STARTTLS (when offered) otherwise.
func (s *SMTP) deliver(to string, msg []byte) error {
	dialer := &net.Dialer{Timeout: dialTimeout}

	if s.cfg.Port == 465 {
		conn, err := tls.DialWithDialer(dialer, "tcp", s.addr, &tls.Config{ServerName: s.cfg.Host})
		if err != nil {
			return fmt.Errorf("mailer: dial: %w", err)
		}
		c, err := smtp.NewClient(conn, s.cfg.Host)
		if err != nil {
			_ = conn.Close()
			return fmt.Errorf("mailer: smtp client: %w", err)
		}
		return s.converse(c, to, msg)
	}

	conn, err := dialer.Dial("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("mailer: dial: %w", err)
	}
	c, err := smtp.NewClient(conn, s.cfg.Host)
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("mailer: smtp client: %w", err)
	}
	if ok, _ := c.Extension("STARTTLS"); ok {
		if err := c.StartTLS(&tls.Config{ServerName: s.cfg.Host}); err != nil {
			_ = c.Close()
			return fmt.Errorf("mailer: starttls: %w", err)
		}
	}
	return s.converse(c, to, msg)
}

// converse authenticates (when credentials are set) and transmits one message.
func (s *SMTP) converse(c *smtp.Client, to string, msg []byte) error {
	defer func() { _ = c.Close() }()

	if s.cfg.Username != "" {
		if ok, _ := c.Extension("AUTH"); ok {
			auth := smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)
			if err := c.Auth(auth); err != nil {
				return fmt.Errorf("mailer: auth: %w", err)
			}
		}
	}
	if err := c.Mail(s.fromAddr); err != nil {
		return fmt.Errorf("mailer: MAIL FROM: %w", err)
	}
	if err := c.Rcpt(to); err != nil {
		return fmt.Errorf("mailer: RCPT TO: %w", err)
	}
	w, err := c.Data()
	if err != nil {
		return fmt.Errorf("mailer: DATA: %w", err)
	}
	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("mailer: write body: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("mailer: close body: %w", err)
	}
	return c.Quit()
}

// build assembles a multipart/alternative MIME message. Bodies are
// quoted-printable so accented Portuguese survives any 7-bit relay.
func (s *SMTP) build(to, subject, text, html string) []byte {
	const boundary = "mirante-alt-boundary-b1f3c7d9"
	var b strings.Builder

	fmt.Fprintf(&b, "From: %s\r\n", s.cfg.From)
	fmt.Fprintf(&b, "To: %s\r\n", to)
	fmt.Fprintf(&b, "Subject: %s\r\n", mime.QEncoding.Encode("utf-8", subject))
	fmt.Fprintf(&b, "Date: %s\r\n", time.Now().UTC().Format(time.RFC1123Z))
	b.WriteString("MIME-Version: 1.0\r\n")
	fmt.Fprintf(&b, "Content-Type: multipart/alternative; boundary=%q\r\n\r\n", boundary)

	writePart(&b, boundary, "text/plain", text)
	writePart(&b, boundary, "text/html", html)
	fmt.Fprintf(&b, "--%s--\r\n", boundary)

	return []byte(b.String())
}

// writePart emits one quoted-printable MIME part.
func writePart(b *strings.Builder, boundary, ctype, body string) {
	fmt.Fprintf(b, "--%s\r\n", boundary)
	fmt.Fprintf(b, "Content-Type: %s; charset=utf-8\r\n", ctype)
	b.WriteString("Content-Transfer-Encoding: quoted-printable\r\n\r\n")

	var enc strings.Builder
	qp := quotedprintable.NewWriter(&enc)
	_, _ = qp.Write([]byte(body))
	_ = qp.Close()

	b.WriteString(enc.String())
	b.WriteString("\r\n")
}
