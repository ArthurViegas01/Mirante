// Package llm is the provider-agnostic LLM gateway (ADR-0004). One provider is
// selected by env (no runtime failover between providers). All calls go through
// Client, which enforces a per-route rate limit and records every call in a usage
// ledger. Structured output is requested as a JSON object and "validated" by
// unmarshaling into the caller's typed result. Free-form input (a job posting,
// résumé text) is always passed as data, never as instructions.
package llm

import (
	"context"
	"errors"
)

// Role is a chat message role.
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

// Message is a single chat turn.
type Message struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
}

// Request is one completion request. System frames the task; User carries the
// (untrusted) data. JSON asks the provider to return a single JSON object.
type Request struct {
	System      string
	User        string
	MaxTokens   int
	Temperature float64
	JSON        bool
}

// Usage is provider-reported token accounting.
type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// Response is a completion result.
type Response struct {
	Content string
	Model   string
	Usage   Usage
}

// Provider is a concrete LLM backend (Groq, Anthropic, OpenAI, mock).
type Provider interface {
	Name() string
	Model() string
	Complete(ctx context.Context, r Request) (*Response, error)
}

// UsageEntry is one ledger row.
type UsageEntry struct {
	Provider     string
	Model        string
	Route        string
	InputTokens  int
	OutputTokens int
}

// Ledger records usage for cost/audit; a no-op or DB-backed impl may be used.
type Ledger interface {
	Record(ctx context.Context, e UsageEntry) error
}

// RouteLimiter caps calls per logical route (owner quota, not per-IP).
type RouteLimiter interface {
	Allow(route string) bool
}

// Errors.
var (
	ErrNoProvider  = errors.New("llm: no provider configured")
	ErrRateLimited = errors.New("llm: route rate limit exceeded")
	ErrProvider    = errors.New("llm: provider error")
)
