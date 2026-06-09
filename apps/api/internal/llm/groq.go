package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	groqBaseURL      = "https://api.groq.com/openai/v1"
	groqDefaultModel = "llama-3.3-70b-versatile"
)

type groqProvider struct {
	apiKey  string
	model   string
	baseURL string
	http    *http.Client
}

// NewGroq builds a Groq provider (OpenAI-compatible Chat Completions). An empty
// model falls back to a sensible default.
func NewGroq(apiKey, model string) Provider {
	return newGroq(apiKey, model, groqBaseURL)
}

func newGroq(apiKey, model, baseURL string) *groqProvider {
	if model == "" {
		model = groqDefaultModel
	}
	return &groqProvider{
		apiKey:  apiKey,
		model:   model,
		baseURL: baseURL,
		http:    &http.Client{Timeout: 60 * time.Second},
	}
}

func (g *groqProvider) Name() string  { return "groq" }
func (g *groqProvider) Model() string { return g.model }

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type responseFormat struct {
	Type string `json:"type"`
}

type chatRequest struct {
	Model          string          `json:"model"`
	Messages       []chatMessage   `json:"messages"`
	MaxTokens      int             `json:"max_tokens,omitempty"`
	Temperature    float64         `json:"temperature,omitempty"`
	ResponseFormat *responseFormat `json:"response_format,omitempty"`
}

type chatResponse struct {
	Model   string `json:"model"`
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (g *groqProvider) Complete(ctx context.Context, r Request) (*Response, error) {
	msgs := make([]chatMessage, 0, 2)
	if r.System != "" {
		msgs = append(msgs, chatMessage{Role: string(RoleSystem), Content: r.System})
	}
	msgs = append(msgs, chatMessage{Role: string(RoleUser), Content: r.User})

	reqBody := chatRequest{
		Model:       g.model,
		Messages:    msgs,
		MaxTokens:   r.MaxTokens,
		Temperature: r.Temperature,
	}
	if r.JSON {
		reqBody.ResponseFormat = &responseFormat{Type: "json_object"}
	}

	buf, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, g.baseURL+"/chat/completions", bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+g.apiKey)

	resp, err := g.http.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrProvider, err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}

	var cr chatResponse
	if err := json.Unmarshal(body, &cr); err != nil {
		return nil, fmt.Errorf("%w: decode response: %w", ErrProvider, err)
	}
	if resp.StatusCode != http.StatusOK {
		msg := resp.Status
		if cr.Error != nil {
			msg = cr.Error.Message
		}
		return nil, fmt.Errorf("%w: %s", ErrProvider, msg)
	}
	if len(cr.Choices) == 0 {
		return nil, fmt.Errorf("%w: empty choices", ErrProvider)
	}

	return &Response{
		Content: cr.Choices[0].Message.Content,
		Model:   cr.Model,
		Usage:   Usage{InputTokens: cr.Usage.PromptTokens, OutputTokens: cr.Usage.CompletionTokens},
	}, nil
}
