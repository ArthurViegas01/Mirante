// Package config loads all runtime configuration from the environment.
// Secrets never have defaults and never appear in code.
package config

import (
	"errors"
	"fmt"
	"os"
	"time"
)

// Config is the fully-resolved runtime configuration.
type Config struct {
	AppEnv    string
	HTTPAddr  string
	WebOrigin string

	DatabaseURL   string
	DatabaseToken string

	SessionCookie string
	SessionTTL    time.Duration

	OwnerEmail    string
	OwnerPassword string
	OwnerHash     string

	SecretKey string

	// E-mail (optional). The password-reset flow uses Resend (HTTP API) when
	// ResendAPIKey is set, else SMTP when SMTPHost is set, else logs the link.
	// MailFrom is the sender for whichever transport is active.
	ResendAPIKey     string
	MailFrom         string
	SMTPHost         string
	SMTPPort         int
	SMTPUsername     string
	SMTPPassword     string
	SMTPFrom         string
	PasswordResetTTL time.Duration

	LLMProvider      string
	LLMModel         string
	LLMAPIKey        string
	LLMRatePerMinute int

	MonitorRetention time.Duration
	AlertWebhookURL  string

	OtelService  string
	OtelEndpoint string
}

// IsProd reports whether the app runs in production mode.
func (c Config) IsProd() bool { return c.AppEnv == "production" }

// Load reads and validates configuration from the environment.
func Load() (Config, error) {
	c := Config{
		AppEnv:   env("APP_ENV", "development"),
		HTTPAddr: httpAddr(),
		WebOrigin:       env("WEB_ORIGIN", "http://localhost:5173"),
		DatabaseURL:     env("DATABASE_URL", "file:./data/mirante.db"),
		DatabaseToken:   env("DATABASE_AUTH_TOKEN", ""),
		SessionCookie:   env("SESSION_COOKIE_NAME", "mirante_session"),
		OwnerEmail:      env("OWNER_EMAIL", ""),
		OwnerPassword:   env("OWNER_PASSWORD", ""),
		OwnerHash:       env("OWNER_PASSWORD_HASH", ""),
		SecretKey:       env("APP_SECRET_KEY", ""),
		ResendAPIKey:    env("RESEND_API_KEY", ""),
		SMTPHost:        env("SMTP_HOST", ""),
		SMTPUsername:    env("SMTP_USERNAME", ""),
		SMTPPassword:    env("SMTP_PASSWORD", ""),
		SMTPFrom:        env("SMTP_FROM", ""),
		LLMProvider:     env("LLM_PROVIDER", "groq"),
		LLMModel:        env("LLM_MODEL", ""),
		OtelService:     env("OTEL_SERVICE_NAME", "mirante-api"),
		OtelEndpoint:    env("OTEL_EXPORTER_OTLP_ENDPOINT", ""),
		AlertWebhookURL: env("ALERT_WEBHOOK_URL", ""),
	}

	// The API key is read generically (LLM_API_KEY) or from the provider's
	// conventional var. Absent in dev → LLM features fall back to a mock.
	c.LLMAPIKey = env("LLM_API_KEY", "")
	if c.LLMAPIKey == "" {
		switch c.LLMProvider {
		case "groq":
			c.LLMAPIKey = env("GROQ_API_KEY", "")
		case "anthropic":
			c.LLMAPIKey = env("ANTHROPIC_API_KEY", "")
		case "openai":
			c.LLMAPIKey = env("OPENAI_API_KEY", "")
		}
	}

	ttl, err := time.ParseDuration(env("SESSION_TTL", "720h"))
	if err != nil {
		return Config{}, fmt.Errorf("invalid SESSION_TTL: %w", err)
	}
	c.SessionTTL = ttl

	rate, err := atoi(env("LLM_RATE_PER_MINUTE", "20"))
	if err != nil {
		return Config{}, fmt.Errorf("invalid LLM_RATE_PER_MINUTE: %w", err)
	}
	c.LLMRatePerMinute = rate

	retentionDays, err := atoi(env("MONITOR_RETENTION_DAYS", "14"))
	if err != nil {
		return Config{}, fmt.Errorf("invalid MONITOR_RETENTION_DAYS: %w", err)
	}
	if retentionDays < 1 {
		return Config{}, errors.New("MONITOR_RETENTION_DAYS must be >= 1")
	}
	c.MonitorRetention = time.Duration(retentionDays) * 24 * time.Hour

	smtpPort, err := atoi(env("SMTP_PORT", "587"))
	if err != nil {
		return Config{}, fmt.Errorf("invalid SMTP_PORT: %w", err)
	}
	c.SMTPPort = smtpPort

	// Canonical sender; MAIL_FROM wins, else the legacy SMTP_FROM.
	c.MailFrom = env("MAIL_FROM", c.SMTPFrom)

	resetTTL, err := time.ParseDuration(env("PASSWORD_RESET_TTL", "1h"))
	if err != nil {
		return Config{}, fmt.Errorf("invalid PASSWORD_RESET_TTL: %w", err)
	}
	c.PasswordResetTTL = resetTTL

	if c.IsProd() {
		if c.SecretKey == "" {
			return Config{}, errors.New("APP_SECRET_KEY is required in production")
		}
		if c.WebOrigin == "" {
			return Config{}, errors.New("WEB_ORIGIN is required in production")
		}
	}

	return c, nil
}

// httpAddr returns the address to listen on, preferring HTTP_ADDR, then PORT
// (Railway / Heroku convention), then defaulting to :8080.
func httpAddr() string {
	if v := os.Getenv("HTTP_ADDR"); v != "" {
		return v
	}
	if p := os.Getenv("PORT"); p != "" {
		return ":" + p
	}
	return ":8080"
}

func env(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

func atoi(s string) (int, error) {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}
