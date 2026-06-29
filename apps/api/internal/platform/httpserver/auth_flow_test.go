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
		RateLimit(ratelimit.New(240, time.Minute), ""),
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

func TestSignupFirstIsAdminThenPending(t *testing.T) {
	base, client := setupNoOwner(t)

	// Fresh instance needs setup, and a protected route is closed.
	require.True(t, needsSetup(t, client, base))
	require.Equal(t, http.StatusUnauthorized,
		statusOf(t, client, newRequest(t, http.MethodGet, base+"/api/auth/me", "", "")))

	// First signup claims the admin and logs in (cookie set, CSRF returned).
	status, csrf := doSignup(t, client, base, testEmail, testPass, "Owner")
	require.Equal(t, http.StatusCreated, status)
	require.NotEmpty(t, csrf)
	require.Equal(t, http.StatusOK,
		statusOf(t, client, newRequest(t, http.MethodGet, base+"/api/auth/me", "", "")))
	require.False(t, needsSetup(t, client, base))

	// Registration stays open, but a later signup is created pending (202, no
	// session) and cannot log in until an admin activates it.
	fresh := newClient(t)
	pending, pendCSRF := doSignup(t, fresh, base, "newuser@example.com", "another-pass", "")
	require.Equal(t, http.StatusAccepted, pending)
	require.Empty(t, pendCSRF)
	notYet, _ := doLogin(t, newClient(t), base, "newuser@example.com", "another-pass")
	require.Equal(t, http.StatusForbidden, notYet)

	// The admin can log in with the credentials chosen at signup.
	loginStatus, _ := doLogin(t, newClient(t), base, testEmail, testPass)
	require.Equal(t, http.StatusOK, loginStatus)
}

func TestForgotPasswordAlwaysOK(t *testing.T) {
	base, client := setup(t) // owner seeded; serve() wires no mailer, so links are logged

	// Known owner address.
	require.Equal(t, http.StatusOK, statusOf(t, client,
		newRequest(t, http.MethodPost, base+"/api/auth/forgot-password",
			fmt.Sprintf(`{"email":%q}`, testEmail), "")))

	// Unknown address — still 200, so the endpoint can't probe for the owner.
	require.Equal(t, http.StatusOK, statusOf(t, client,
		newRequest(t, http.MethodPost, base+"/api/auth/forgot-password",
			`{"email":"stranger@example.com"}`, "")))
}

func TestForgotPasswordRejectsInvalidEmail(t *testing.T) {
	base, client := setup(t)
	require.Equal(t, http.StatusBadRequest, statusOf(t, client,
		newRequest(t, http.MethodPost, base+"/api/auth/forgot-password",
			`{"email":"not-an-email"}`, "")))
}

func TestForgotPasswordRejectsForeignOrigin(t *testing.T) {
	base, client := setup(t)
	req := newRequest(t, http.MethodPost, base+"/api/auth/forgot-password",
		fmt.Sprintf(`{"email":%q}`, testEmail), "")
	req.Header.Set("Origin", "http://evil.example")
	require.Equal(t, http.StatusForbidden, statusOf(t, client, req))
}

func TestResetPasswordRejectsBadToken(t *testing.T) {
	base, client := setup(t)
	require.Equal(t, http.StatusBadRequest, statusOf(t, client,
		newRequest(t, http.MethodPost, base+"/api/auth/reset-password",
			`{"token":"bogus-token","password":"new-password-123"}`, "")))
}

func TestSignupRejectsShortPassword(t *testing.T) {
	base, client := setupNoOwner(t)
	status, _ := doSignup(t, client, base, testEmail, "short", "")
	require.Equal(t, http.StatusBadRequest, status)
	require.True(t, needsSetup(t, client, base)) // nothing was created
}

func TestSignupWhenAdminExistsIsPending(t *testing.T) {
	base, client := setup(t) // admin seeded from env-style bootstrap
	require.False(t, needsSetup(t, client, base))
	status, _ := doSignup(t, client, base, "someone@example.com", "a-valid-pass", "")
	require.Equal(t, http.StatusAccepted, status) // created, awaiting activation

	// A duplicate e-mail is rejected.
	dup, _ := doSignup(t, newClient(t), base, testEmail, "whatever-pass", "")
	require.Equal(t, http.StatusConflict, dup)
}

func TestAdminActivationFlow(t *testing.T) {
	base, admin := setup(t) // testEmail is the seeded admin (active)
	_, csrf := doLogin(t, admin, base, testEmail, testPass)
	require.NotEmpty(t, csrf)

	// A stranger signs up → pending, cannot log in.
	st, _ := doSignup(t, newClient(t), base, "pend@example.com", "pending-pass", "Pend")
	require.Equal(t, http.StatusAccepted, st)
	blocked, _ := doLogin(t, newClient(t), base, "pend@example.com", "pending-pass")
	require.Equal(t, http.StatusForbidden, blocked)

	// Admin finds the pending account and activates it.
	resp, err := admin.Do(newRequest(t, http.MethodGet, base+"/api/admin/users", "", ""))
	require.NoError(t, err)
	var lb struct {
		Users []struct{ ID, Email, Role, Status string } `json:"users"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&lb))
	_ = resp.Body.Close()
	var pendID string
	for _, u := range lb.Users {
		if u.Email == "pend@example.com" {
			require.Equal(t, "user", u.Role)
			require.Equal(t, "pending", u.Status)
			pendID = u.ID
		}
	}
	require.NotEmpty(t, pendID)

	require.Equal(t, http.StatusOK, statusOf(t, admin,
		newRequest(t, http.MethodPost, base+"/api/admin/users/"+pendID+"/activate", "", csrf)))

	// Now the user can log in — and, as a regular user, is denied the admin API.
	pendClient := newClient(t)
	ok, _ := doLogin(t, pendClient, base, "pend@example.com", "pending-pass")
	require.Equal(t, http.StatusOK, ok)
	require.Equal(t, http.StatusForbidden,
		statusOf(t, pendClient, newRequest(t, http.MethodGet, base+"/api/admin/users", "", "")))

	// An anonymous caller is unauthenticated, not just forbidden.
	require.Equal(t, http.StatusUnauthorized,
		statusOf(t, newClient(t), newRequest(t, http.MethodGet, base+"/api/admin/users", "", "")))
}
