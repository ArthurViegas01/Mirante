// Package httpserver assembles the HTTP router, the middleware stack, and the
// auth routes. Domain modules (projects, monitor, …) will register their own
// routers here from F1 onward.
package httpserver

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/lumni/mirante/internal/platform/auth"
	"github.com/lumni/mirante/internal/platform/config"
	"github.com/lumni/mirante/internal/platform/ratelimit"
)

// Deps are the dependencies the router needs.
type Deps struct {
	Log    *slog.Logger
	Auth   *auth.Service
	Config config.Config
}

// Router builds the fully-wired HTTP handler.
func Router(d Deps) http.Handler {
	ah := NewAuthHandlers(d.Auth, AuthConfig{
		CookieName:    d.Config.SessionCookie,
		Secure:        d.Config.IsProd(),
		AllowedOrigin: d.Config.WebOrigin,
	})

	mux := http.NewServeMux()

	// Public.
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.Handle("POST /api/auth/login", http.HandlerFunc(ah.Login))

	// Authenticated.
	mux.Handle("GET /api/auth/me", ah.RequireAuth(http.HandlerFunc(ah.Me)))
	mux.Handle("POST /api/auth/logout", ah.RequireAuth(ah.CSRF(http.HandlerFunc(ah.Logout))))
	// Sample protected route proving the session gate works end to end.
	mux.Handle("GET /api/ping", ah.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"pong": "ok"})
	})))

	ipLimiter := ratelimit.New(240, time.Minute)

	return chain(mux,
		RequestID(),
		Recover(d.Log),
		SecurityHeaders(d.Config.IsProd()),
		CORS(d.Config.WebOrigin),
		RateLimit(ipLimiter),
	)
}
