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
	"github.com/lumni/mirante/internal/platform/db"
	"github.com/lumni/mirante/internal/platform/migrate"
	"github.com/lumni/mirante/internal/platform/ratelimit"
)

const (
	testEmail  = "owner@example.com"
	testPass   = "s3cret-password"
	testOrigin = "http://localhost:5173"
)

func setup(t *testing.T) (string, *http.Client) { return serve(t, true) }

// setupNoOwner builds the server with no owner bootstrapped, so the first-run
// signup flow is open.
func setupNoOwner(t *testing.T) (string, *http.Client) { return serve(t, false) }

func serve(t *testing.T, bootstrap bool) (string, *http.Client) {
	t.Helper()
	ctx := context.Background()

	database, err := db.Open(ctx, ":memory:", "")
	require.NoError(t, err)
	t.Cleanup(func() { _ = database.Close() })
	require.NoError(t, migrate.Up(database.DB))

	svc := auth.NewService(database.DB, time.Hour)
	if bootstrap {
		require.NoError(t, svc.Bootstrap(ctx, testEmail, testPass, ""))
	}

	authH := NewAuthHandlers(svc, AuthConfig{
		CookieName:    "mirante_session",
		Secure:        false,
		AllowedOrigin: testOrigin,
	})

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", Healthz)
	authH.RegisterRoutes(mux)

	handler := Chain(mux,
		RequestID(),
		Recover(slog.New(slog.NewTextHandler(io.Discard, nil))),
		SecurityHeaders(false),
		CORS(testOrigin),
		RateLimit(ratelimit.New(240, time.Minute)),
	)

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

func doSignup(t *testing.T, client *http.Client, base, email, pass, name string) (int, string) {
	t.Helper()
	body := fmt.Sprintf(`{"email":%q,"password":%q,"name":%q}`, email, pass, name)
	resp, err := client.Do(newRequest(t, http.MethodPost, base+"/api/auth/signup", body, ""))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	var b struct {
		CSRFToken string `json:"csrf_token"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&b)
	return resp.StatusCode, b.CSRFToken
}

func newClient(t *testing.T) *http.Client {
	t.Helper()
	jar, err := cookiejar.New(nil)
	require.NoError(t, err)
	return &http.Client{Jar: jar}
}

func needsSetup(t *testing.T, client *http.Client, base string) bool {
	t.Helper()
	resp, err := client.Do(newRequest(t, http.MethodGet, base+"/api/auth/status", "", ""))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	var b struct {
		NeedsSetup bool `json:"needs_setup"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&b))
	return b.NeedsSetup
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

func TestSignupClaimsInstanceThenCloses(t *testing.T) {
	base, client := setupNoOwner(t)

	// Fresh instance needs setup, and a protected route is closed.
	require.True(t, needsSetup(t, client, base))
	require.Equal(t, http.StatusUnauthorized,
		statusOf(t, client, newRequest(t, http.MethodGet, base+"/api/auth/me", "", "")))

	// Signup claims the owner and logs in (cookie set, CSRF returned).
	status, csrf := doSignup(t, client, base, testEmail, testPass, "Owner")
	require.Equal(t, http.StatusCreated, status)
	require.NotEmpty(t, csrf)
	require.Equal(t, http.StatusOK,
		statusOf(t, client, newRequest(t, http.MethodGet, base+"/api/auth/me", "", "")))

	// Setup is now done and registration is closed for a new visitor.
	require.False(t, needsSetup(t, client, base))
	fresh := newClient(t)
	closed, _ := doSignup(t, fresh, base, "intruder@example.com", "another-pass", "")
	require.Equal(t, http.StatusForbidden, closed)

	// The owner can log in with the credentials chosen at signup.
	loginStatus, _ := doLogin(t, fresh, base, testEmail, testPass)
	require.Equal(t, http.StatusOK, loginStatus)
}

func TestSignupRejectsShortPassword(t *testing.T) {
	base, client := setupNoOwner(t)
	status, _ := doSignup(t, client, base, testEmail, "short", "")
	require.Equal(t, http.StatusBadRequest, status)
	require.True(t, needsSetup(t, client, base)) // nothing was created
}

func TestSignupClosedWhenOwnerBootstrapped(t *testing.T) {
	base, client := setup(t) // owner seeded from env-style bootstrap
	require.False(t, needsSetup(t, client, base))
	status, _ := doSignup(t, client, base, "someone@example.com", "a-valid-pass", "")
	require.Equal(t, http.StatusForbidden, status)
}
