package llm

// NewProvider builds the configured provider. It returns (nil, false) when no API
// key is set or the provider is not implemented, so the composition root can fall
// back to a mock (or leave the Client unavailable) and the app still boots.
//
// Only Groq is implemented today (ADR-0004: one provider, chosen by env). Adding
// Anthropic/OpenAI means a new case here plus a sibling of groq.go.
func NewProvider(name, model, apiKey string) (Provider, bool) {
	if apiKey == "" {
		return nil, false
	}
	switch name {
	case "groq":
		return NewGroq(apiKey, model), true
	default:
		return nil, false
	}
}
