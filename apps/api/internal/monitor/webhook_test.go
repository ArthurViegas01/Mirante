package monitor

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func sampleAlert() Alert {
	return Alert{
		ID:         7,
		ServiceID:  "svc1",
		ProjectID:  "proj1",
		Severity:   "danger",
		Title:      "API está fora do ar",
		Body:       "no_response",
		FromStatus: StatusUp,
		ToStatus:   StatusDown,
		CreatedAt:  time.Now().UTC(),
	}
}

func TestWebhookChannelSendsJSON(t *testing.T) {
	var got webhookPayload
	var gotCT string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		gotCT = r.Header.Get("Content-Type")
		require.NoError(t, json.NewDecoder(r.Body).Decode(&got))
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	ch, err := NewWebhookChannel(srv.URL, nil)
	require.NoError(t, err)
	require.Equal(t, "webhook", ch.Name())
	require.NoError(t, ch.Send(context.Background(), sampleAlert()))

	require.Equal(t, "application/json", gotCT)
	require.Equal(t, "monitor.transition", got.Event)
	require.Equal(t, int64(7), got.AlertID)
	require.Equal(t, "svc1", got.ServiceID)
	require.Equal(t, "danger", got.Severity)
	require.Equal(t, "API está fora do ar", got.Title)
	require.Equal(t, "down", got.ToStatus)
	require.Equal(t, "up", got.FromStatus)
	require.NotEmpty(t, got.At)
}

func TestWebhookChannelNon2xxIsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	ch, err := NewWebhookChannel(srv.URL, nil)
	require.NoError(t, err)
	require.Error(t, ch.Send(context.Background(), sampleAlert()))
}

func TestNewWebhookChannelRejectsBadURL(t *testing.T) {
	for _, bad := range []string{"", "ftp://example.com", "not-a-url", "https://"} {
		_, err := NewWebhookChannel(bad, nil)
		require.Error(t, err, "url %q should be rejected", bad)
	}
	_, err := NewWebhookChannel("https://hooks.example.com/abc", nil)
	require.NoError(t, err)
}

// The Notifier isolates a failing channel: Dispatch must not panic or block when
// the endpoint errors (the error is logged, not propagated).
func TestNotifierDispatchToWebhook(t *testing.T) {
	hits := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	ch, err := NewWebhookChannel(srv.URL, nil)
	require.NoError(t, err)
	n := NewNotifier(discardLog(), ch)
	n.Dispatch(context.Background(), sampleAlert())
	require.Equal(t, 1, hits)
}
