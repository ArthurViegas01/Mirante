package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	idb "github.com/lumni/mirante/internal/platform/db"
	"github.com/lumni/mirante/internal/platform/migrate"
)

func TestGroqComplete(t *testing.T) {
	var gotAuth string
	var gotBody chatRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"model":"llama-test","choices":[{"message":{"content":"oi"}}],"usage":{"prompt_tokens":12,"completion_tokens":3}}`))
	}))
	defer srv.Close()

	g := newGroq("sk-test", "llama-test", srv.URL)
	resp, err := g.Complete(context.Background(), Request{System: "sys", User: "hello", JSON: true})
	require.NoError(t, err)
	require.Equal(t, "oi", resp.Content)
	require.Equal(t, "llama-test", resp.Model)
	require.Equal(t, 12, resp.Usage.InputTokens)
	require.Equal(t, 3, resp.Usage.OutputTokens)
	require.Equal(t, "Bearer sk-test", gotAuth)
	require.NotNil(t, gotBody.ResponseFormat)
	require.Equal(t, "json_object", gotBody.ResponseFormat.Type)
	require.Len(t, gotBody.Messages, 2) // system + user
}

func TestGroqError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":{"message":"rate limited"}}`))
	}))
	defer srv.Close()
	g := newGroq("k", "", srv.URL)
	_, err := g.Complete(context.Background(), Request{User: "x"})
	require.ErrorIs(t, err, ErrProvider)
}

type fakeLedger struct{ entries []UsageEntry }

func (f *fakeLedger) Record(_ context.Context, e UsageEntry) error {
	f.entries = append(f.entries, e)
	return nil
}

func TestClientRecordsUsage(t *testing.T) {
	led := &fakeLedger{}
	c := NewClient(NewMock("ok"), led, NewRouteLimiter(0))
	require.True(t, c.Available())

	resp, err := c.Complete(context.Background(), "aderencia", Request{User: "abcd"})
	require.NoError(t, err)
	require.Equal(t, "ok", resp.Content)
	require.Len(t, led.entries, 1)
	require.Equal(t, "mock", led.entries[0].Provider)
	require.Equal(t, "aderencia", led.entries[0].Route)
}

func TestClientNoProvider(t *testing.T) {
	c := NewClient(nil, nil, nil)
	require.False(t, c.Available())
	_, err := c.Complete(context.Background(), "r", Request{})
	require.ErrorIs(t, err, ErrNoProvider)
}

func TestClientRateLimited(t *testing.T) {
	c := NewClient(NewMock("x"), nil, NewRouteLimiter(1))
	_, err := c.Complete(context.Background(), "r", Request{})
	require.NoError(t, err)
	_, err = c.Complete(context.Background(), "r", Request{})
	require.ErrorIs(t, err, ErrRateLimited)
	// A different route has its own window.
	_, err = c.Complete(context.Background(), "other", Request{})
	require.NoError(t, err)
}

func TestCompleteJSON(t *testing.T) {
	c := NewClient(NewMock(`{"score":80,"skills":["Go"]}`), nil, nil)
	var out struct {
		Score  int      `json:"score"`
		Skills []string `json:"skills"`
	}
	require.NoError(t, c.CompleteJSON(context.Background(), "r", Request{User: "j"}, &out))
	require.Equal(t, 80, out.Score)
	require.Equal(t, []string{"Go"}, out.Skills)

	bad := NewClient(NewMock("not json"), nil, nil)
	require.ErrorIs(t, bad.CompleteJSON(context.Background(), "r", Request{}, &out), ErrProvider)
}

func TestSQLiteLedger(t *testing.T) {
	ctx := context.Background()
	database, err := idb.Open(ctx, ":memory:", "")
	require.NoError(t, err)
	t.Cleanup(func() { _ = database.Close() })
	require.NoError(t, migrate.Up(database.DB))

	led := NewSQLiteLedger(database)
	require.NoError(t, led.Record(ctx, UsageEntry{
		Provider: "groq", Model: "m", Route: "aderencia", InputTokens: 10, OutputTokens: 5,
	}))

	var count, sumIn int
	require.NoError(t, database.QueryRowContext(ctx,
		`SELECT COUNT(*), COALESCE(SUM(input_tokens),0) FROM llm_usage`).Scan(&count, &sumIn))
	require.Equal(t, 1, count)
	require.Equal(t, 10, sumIn)
}

func TestProviderFactory(t *testing.T) {
	_, ok := NewProvider("groq", "", "")
	require.False(t, ok) // no key → unconfigured

	p, ok := NewProvider("groq", "m", "key")
	require.True(t, ok)
	require.Equal(t, "groq", p.Name())

	_, ok = NewProvider("anthropic", "", "key")
	require.False(t, ok) // not implemented yet
}
