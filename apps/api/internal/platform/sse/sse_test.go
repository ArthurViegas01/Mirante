package sse

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/lumni/mirante/internal/platform/tenant"
)

func TestHubReplayAndLive(t *testing.T) {
	// replay returns one event with id = after+1.
	hub := NewHub(func(_ context.Context, after int64, _ int) ([]Event, error) {
		return []Event{{ID: after + 1, Type: "monitor.transition", Data: []byte(`{"x":1}`)}}, nil
	})
	srv := httptest.NewServer(hub)
	defer srv.Close()

	req, err := http.NewRequest(http.MethodGet, srv.URL, nil)
	require.NoError(t, err)
	req.Header.Set("Last-Event-ID", "5")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, "text/event-stream", resp.Header.Get("Content-Type"))

	// Wait for the client to register, then publish a live event.
	require.Eventually(t, func() bool { return hub.Clients() == 1 }, time.Second, 5*time.Millisecond)
	hub.Emit("", 7, "monitor.transition", []byte(`{"x":2}`))

	got := readFor(resp.Body, 250*time.Millisecond)
	require.Contains(t, got, "id: 6") // replayed (after 5)
	require.Contains(t, got, "id: 7") // live
}

func TestHubRejectsTooManyClients(t *testing.T) {
	hub := NewHub(nil)
	hub.maxClients = 0
	srv := httptest.NewServer(hub)
	defer srv.Close()

	resp, err := http.Get(srv.URL)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
}

// tenantHandler injects an owner into the request context, standing in for the
// auth middleware so the hub can scope a connection to a user.
func tenantHandler(uid string, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r.WithContext(tenant.WithUserID(r.Context(), uid)))
	})
}

// TestHubPerUserFanout: a live event reaches only its owner's connection, never
// another user's stream.
func TestHubPerUserFanout(t *testing.T) {
	hub := NewHub(nil)
	srvA := httptest.NewServer(tenantHandler("user-a", hub))
	defer srvA.Close()
	srvB := httptest.NewServer(tenantHandler("user-b", hub))
	defer srvB.Close()

	respA, err := http.Get(srvA.URL)
	require.NoError(t, err)
	defer func() { _ = respA.Body.Close() }()
	respB, err := http.Get(srvB.URL)
	require.NoError(t, err)
	defer func() { _ = respB.Body.Close() }()

	require.Eventually(t, func() bool { return hub.Clients() == 2 }, time.Second, 5*time.Millisecond)
	hub.Emit("user-a", 9, "monitor.transition", []byte(`{"x":1}`))

	gotA := readFor(respA.Body, 250*time.Millisecond)
	gotB := readFor(respB.Body, 250*time.Millisecond)
	require.Contains(t, gotA, "id: 9")    // the owner receives it
	require.NotContains(t, gotB, "id: 9") // the other user never does
}

// readFor accumulates streamed bytes for d, then closes the body to unblock.
func readFor(body io.ReadCloser, d time.Duration) string {
	var mu sync.Mutex
	var sb strings.Builder
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := body.Read(buf)
			if n > 0 {
				mu.Lock()
				sb.Write(buf[:n])
				mu.Unlock()
			}
			if err != nil {
				return
			}
		}
	}()
	time.Sleep(d)
	_ = body.Close()
	time.Sleep(10 * time.Millisecond)
	mu.Lock()
	defer mu.Unlock()
	return sb.String()
}
