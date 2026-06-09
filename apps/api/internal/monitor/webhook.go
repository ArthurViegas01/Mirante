package monitor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	idb "github.com/lumni/mirante/internal/platform/db"
)

// WebhookChannel delivers transition alerts to an owner-configured HTTP endpoint
// as JSON. The URL is set by the owner via env (trusted), so — unlike job-link
// import (ADR-0003) — it is not behind the SSRF guard: the owner points it at
// their own receiver (Discord/Slack relay, a personal endpoint, etc.).
type WebhookChannel struct {
	url    string
	client *http.Client
}

// NewWebhookChannel builds a webhook channel for an http(s) URL. A nil client
// gets a default with a 10s timeout. Returns an error if the URL is not http(s).
func NewWebhookChannel(rawURL string, client *http.Client) (*WebhookChannel, error) {
	u, err := url.Parse(rawURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return nil, fmt.Errorf("webhook url must be an http(s) URL")
	}
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &WebhookChannel{url: rawURL, client: client}, nil
}

// Name identifies the channel in logs.
func (c *WebhookChannel) Name() string { return "webhook" }

// webhookPayload is the JSON posted on each transition. Title is human-readable
// and already includes the service name; the target/credentials never appear.
type webhookPayload struct {
	Event      string `json:"event"`
	AlertID    int64  `json:"alert_id"`
	ServiceID  string `json:"service_id"`
	ProjectID  string `json:"project_id"`
	Severity   string `json:"severity"`
	Title      string `json:"title"`
	Body       string `json:"body,omitempty"`
	FromStatus string `json:"from_status"`
	ToStatus   string `json:"to_status"`
	At         string `json:"at"`
}

// Send posts the alert as JSON, honoring the context deadline. A non-2xx
// response (or transport error) is returned so the Notifier can log it.
func (c *WebhookChannel) Send(ctx context.Context, a Alert) error {
	body, err := json.Marshal(webhookPayload{
		Event:      "monitor.transition",
		AlertID:    a.ID,
		ServiceID:  string(a.ServiceID),
		ProjectID:  a.ProjectID,
		Severity:   a.Severity,
		Title:      a.Title,
		Body:       a.Body,
		FromStatus: string(a.FromStatus),
		ToStatus:   string(a.ToStatus),
		At:         idb.FormatTime(a.CreatedAt),
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	_, _ = io.Copy(io.Discard, resp.Body)

	if resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}
	return nil
}
