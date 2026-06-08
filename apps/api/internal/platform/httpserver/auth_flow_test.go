package httpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/lumni/mirante/internal/platform/auth"
	"github.com/lumni/mirante/internal/platform/config"
	"github.com/lumni/mirante/internal/platform/db"
	"github.com/lumni/mirante/internal/platform/migrate"
)

const (
	testEmail = "owner@example.com"
	testPass  = "s3cret-password"
)

func setup(t *testing.T) (string, *http.Client) {
	t.Helper()
	ctx := context.Background()

	database, err := db.Open(ctx, ":memory:", "")
	require.NoError(t, err)
	t.Cleanup(func() { _ = database.Close() })
	require.NoError(t, migrate.Up(database.DB))

	svc := auth.NewService(database.DB, time.Hour)
	require.NoError(t, svc.Bootstrap(ctx, testEmail, testPass, ""))

	cfg := config.Config{
		AppEnv:        "development",
		WebOrigin:     "http://localhost:5173",
		SessionCookie: "mirante_session",
		SessionTTL:    time.Hour,
	}
	handler := Router(Deps{
		Log:    slog.New(slog.NewTextHandler(io.Discard, nil)),
		Auth:   svc,
		Config: cfg,
	})

	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	jar, err := cookiejar.New(nil)
	require.NoError(t, err)
	return srv.URL, &http.Client{Jar: jar}
}

func newRequest(t *testing.T, method, url, body, csrf string) *http.Request {
	t.Helper()
	var r io.Reader
	if body != "" {
		r = strings.NewReader(body)
	}
	req, err := http.NewRequest(method, url, r)
	require.NoError(t, err)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if csrf != "" {
		req.Header.Set("X-CSRF-Token", csrf)
	}
	return req
}

func doLogin(t *testing.T, client *http.Client, base, email, pass string) (int, string) {
	t.Helper()
	body := fmt.Sprintf(`{"email":%q,"password":%q}`, email, pass)
	resp, err := client.Do(newRequest(t, http.MethodPost, base+"/api/auth/login", body, ""))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	var b struct {
		CSRFToken string `json:"csrf_token"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&b)
	return resp.StatusCode, b.CSRFToken
}

func statusOf(t *testing.T, client *http.Client, req *http.Request) int {
	t.Helper()
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	_, _ = io.Copy(io.Discard, resp.Body)
	return resp.StatusCode
}

func TestAuthFlow(t *testing.T) {
	base, client := setup(t)

	status, csrf := doLogin(t, client, base, testEmail, testPass)
	require.Equal(t, http.StatusOK, status)
	require.NotEmpty(t, csrf)

	require.Equal(t, http.StatusOK,
		statusOf(t, client, newRequest(t, http.MethodGet, base+"/api/auth/me", "", "")))
	require.Equal(t, http.StatusOK,
		statusOf(t, client, newRequest(t, http.MethodGet, base+"/api/ping", "", "")))

	// Logout without a CSRF token must be rejected.
	require.Equal(t, http.StatusForbidden,
		statusOf(t, client, newRequest(t, http.MethodPost, base+"/api/auth/logout", "", "")))

	// Logout with the CSRF token succeeds.
	require.Equal(t, http.StatusOK,
		statusOf(t, client, newRequest(t, http.MethodPost, base+"/api/auth/logout", "", csrf)))

	// After logout the session is gone.
	require.Equal(t, http.StatusUnauthorized,
		statusOf(t, client, newRequest(t, http.MethodGet, base+"/api/auth/me", "", "")))
}

func TestProtectedRouteRequiresSession(t *testing.T) {
	base, client := setup(t)
	require.Equal(t, http.StatusUnauthorized,
		statusOf(t, client, newRequest(t, http.MethodGet, base+"/api/auth/me", "", "")))
}

func TestInvalidCredentials(t *testing.T) {
	base, client := setup(t)
	status, _ := doLogin(t, client, base, testEmail, "wrong")
	require.Equal(t, http.StatusUnauthorized, status)
}

func TestLoginRateLimited(t *testing.T) {
	base, client := setup(t)
	for i := 0; i < 5; i++ {
		status, _ := doLogin(t, client, base, testEmail, "wrong")
		require.Equal(t, http.StatusUnauthorized, status, "attempt %d", i+1)
	}
	status, _ := doLogin(t, client, base, testEmail, "wrong")
	require.Equal(t, http.StatusTooManyRequests, status, "6th attempt should be rate limited")
}

func TestLoginRejectsForeignOrigin(t *testing.T) {
	base, client := setup(t)
	req := newRequest(t, http.MethodPost, base+"/api/auth/login",
		fmt.Sprintf(`{"email":%q,"password":%q}`, testEmail, testPass), "")
	req.Header.Set("Origin", "http://evil.example")
	require.Equal(t, http.StatusForbidden, statusOf(t, client, req))
}
