package llm

import "context"

type mockProvider struct{ reply string }

// NewMock returns a provider that always replies with `reply` (deterministic).
// It is the dev fallback when no API key is configured, and is handy for testing
// consumers (give it the canned JSON your code expects).
func NewMock(reply string) Provider { return &mockProvider{reply: reply} }

func (m *mockProvider) Name() string  { return "mock" }
func (m *mockProvider) Model() string { return "mock" }

func (m *mockProvider) Complete(_ context.Context, r Request) (*Response, error) {
	return &Response{
		Content: m.reply,
		Model:   "mock",
		Usage: Usage{
			InputTokens:  estimateTokens(r.System) + estimateTokens(r.User),
			OutputTokens: estimateTokens(m.reply),
		},
	}, nil
}

// estimateTokens is a rough 4-chars-per-token heuristic for the mock's ledger.
func estimateTokens(s string) int { return (len(s) + 3) / 4 }
