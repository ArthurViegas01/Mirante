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

	LLMProvider      string
	LLMRatePerMinute int

	OtelService  string
	OtelEndpoint string
}

// IsProd reports whether the app runs in production mode.
func (c Config) IsProd() bool { return c.AppEnv == "production" }

// Load reads and validates configuration from the environment.
func Load() (Config, error) {
	c := Config{
		AppEnv:        env("APP_ENV", "development"),
		HTTPAddr:      env("HTTP_ADDR", ":8080"),
		WebOrigin:     env("WEB_ORIGIN", "http://localhost:5173"),
		DatabaseURL:   env("DATABASE_URL", "file:./data/mirante.db"),
		DatabaseToken: env("DATABASE_AUTH_TOKEN", ""),
		SessionCookie: env("SESSION_COOKIE_NAME", "mirante_session"),
		OwnerEmail:    env("OWNER_EMAIL", ""),
		OwnerPassword: env("OWNER_PASSWORD", ""),
		OwnerHash:     env("OWNER_PASSWORD_HASH", ""),
		SecretKey:     env("APP_SECRET_KEY", ""),
		LLMProvider:   env("LLM_PROVIDER", "anthropic"),
		OtelService:   env("OTEL_SERVICE_NAME", "mirante-api"),
		OtelEndpoint:  env("OTEL_EXPORTER_OTLP_ENDPOINT", ""),
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
