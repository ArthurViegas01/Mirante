package mailer

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const resendEndpoint = "https://api.resend.com/emails"

// Resend sends transactional mail via the Resend HTTP API (https://resend.com).
// Preferred over SMTP for app e-mail: better deliverability (the provider handles
// SPF/DKIM/DMARC) and nothing to open on the network — just an HTTPS POST. The
// `from` address must live on a domain verified in Resend (or onboarding@resend.dev
// for testing, which only delivers to your own account address).
type Resend struct {
	apiKey   string
	from     string
	endpoint string
	http     *http.Client
}

// NewResend builds a Resend mailer.
func NewResend(apiKey, from string) (*Resend, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, errors.New("mailer: RESEND_API_KEY is required")
	}
	if strings.TrimSpace(from) == "" {
		return nil, errors.New("mailer: a sender address (MAIL_FROM) is required")
	}
	return &Resend{
		apiKey:   apiKey,
		from:     from,
		endpoint: resendEndpoint,
		http:     &http.Client{Timeout: 15 * time.Second},
	}, nil
}

type resendRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html,omitempty"`
	Text    string   `json:"text,omitempty"`
}

// Send delivers a text+HTML message to a single recipient via the Resend API.
func (r *Resend) Send(ctx context.Context, to, subject, text, html string) error {
	// SetEscapeHTML(false): keep raw <, >, & in the JSON (cleaner, valid JSON that
	// Resend decodes the same) instead of <-escaping the HTML body and sender.
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(resendRequest{
		From: r.from, To: []string{to}, Subject: subject, HTML: html, Text: text,
	}); err != nil {
		return fmt.Errorf("mailer: marshal: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.endpoint, &buf)
	if err != nil {
		return fmt.Errorf("mailer: request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+r.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.http.Do(req)
	if err != nil {
		return fmt.Errorf("mailer: resend send: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 300 {
		// Surface the provider's error (e.g. "domain is not verified") for the log.
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
		return fmt.Errorf("mailer: resend status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4<<10))
	return nil
}
