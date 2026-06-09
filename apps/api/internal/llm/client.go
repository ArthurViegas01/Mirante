package llm

import (
	"context"
	"encoding/json"
	"fmt"
)

// Client is the gateway consumers use. It is safe to construct with a nil
// provider (Available() reports false and Complete returns ErrNoProvider), so the
// app boots without an API key — LLM-backed features simply degrade.
type Client struct {
	provider Provider
	ledger   Ledger
	limiter  RouteLimiter
}

// NewClient wires a provider with an optional ledger and route limiter.
func NewClient(p Provider, ledger Ledger, limiter RouteLimiter) *Client {
	return &Client{provider: p, ledger: ledger, limiter: limiter}
}

// Available reports whether a provider is configured.
func (c *Client) Available() bool { return c != nil && c.provider != nil }

// Complete runs one request on the given route (used for rate limiting and the
// ledger). Usage is recorded best-effort even though the response is returned.
func (c *Client) Complete(ctx context.Context, route string, r Request) (*Response, error) {
	if !c.Available() {
		return nil, ErrNoProvider
	}
	if c.limiter != nil && !c.limiter.Allow(route) {
		return nil, ErrRateLimited
	}
	resp, err := c.provider.Complete(ctx, r)
	if err != nil {
		return nil, err
	}
	if c.ledger != nil {
		_ = c.ledger.Record(ctx, UsageEntry{
			Provider:     c.provider.Name(),
			Model:        resp.Model,
			Route:        route,
			InputTokens:  resp.Usage.InputTokens,
			OutputTokens: resp.Usage.OutputTokens,
		})
	}
	return resp, nil
}

// CompleteJSON requests a JSON object and unmarshals it into dst. A non-JSON or
// schema-mismatched reply surfaces as an error (the "re-validation in Go" step).
func (c *Client) CompleteJSON(ctx context.Context, route string, r Request, dst any) error {
	r.JSON = true
	resp, err := c.Complete(ctx, route, r)
	if err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(resp.Content), dst); err != nil {
		return fmt.Errorf("%w: invalid JSON output: %w", ErrProvider, err)
	}
	return nil
}
