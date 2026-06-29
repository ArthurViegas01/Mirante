package httpserver

import (
	"log/slog"
	"net"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/lumni/mirante/internal/platform/id"
	"github.com/lumni/mirante/internal/platform/ratelimit"
)

// Middleware wraps an http.Handler.
type Middleware func(http.Handler) http.Handler

// Chain applies middlewares so that mws[0] is the outermost wrapper.
func Chain(h http.Handler, mws ...Middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

// RequestID assigns or propagates an X-Request-ID and stores it in the context.
func RequestID() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rid := r.Header.Get("X-Request-ID")
			if rid == "" {
				rid = id.New()
			}
			w.Header().Set("X-Request-ID", rid)
			next.ServeHTTP(w, r.WithContext(withRequestID(r.Context(), rid)))
		})
	}
}

// Recover turns panics into 500s and logs them with the request id.
func Recover(log *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					log.Error("panic recovered",
						"request_id", RequestIDFrom(r.Context()),
						"panic", rec,
						"stack", string(debug.Stack()))
					writeError(w, http.StatusInternalServerError, "internal", "internal error")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// SecurityHeaders sets conservative security headers for API responses.
func SecurityHeaders(isProd bool) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := w.Header()
			h.Set("X-Content-Type-Options", "nosniff")
			h.Set("X-Frame-Options", "DENY")
			h.Set("Referrer-Policy", "no-referrer")
			h.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'; base-uri 'none'")
			h.Set("Cross-Origin-Opener-Policy", "same-origin")
			h.Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
			if isProd {
				h.Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
			}
			next.ServeHTTP(w, r)
		})
	}
}

// CORS restricts cross-origin access to the configured web origin and supports
// credentialed requests (the session cookie).
func CORS(allowedOrigin string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" && origin == allowedOrigin {
				h := w.Header()
				h.Set("Access-Control-Allow-Origin", allowedOrigin)
				h.Set("Access-Control-Allow-Credentials", "true")
				h.Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
				h.Set("Access-Control-Allow-Headers", "Content-Type, X-CSRF-Token, X-Request-ID")
				h.Set("Access-Control-Max-Age", "600")
				h.Add("Vary", "Origin")
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RateLimit throttles requests per client IP. trustedHeader is the proxy header
// to read the real client IP from (empty = ignore proxy headers; see clientIP).
func RateLimit(l *ratelimit.Limiter, trustedHeader string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !l.Allow(clientIP(r, trustedHeader)) {
				writeError(w, http.StatusTooManyRequests, "rate_limited", "too many requests")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// clientIP returns the best-effort client IP. When trustedHeader is non-empty
// (set only when a trusted proxy fronts the app, e.g. Railway/Fly), its value is
// used; otherwise the header is ignored — a directly-exposed instance must not
// trust a forgeable header (F4), so it falls back to the connection's remote
// address. A trustedHeader value with multiple comma-separated entries is read
// left-most; a stray :port is stripped.
func clientIP(r *http.Request, trustedHeader string) string {
	if trustedHeader != "" {
		if v := r.Header.Get(trustedHeader); v != "" {
			if i := strings.IndexByte(v, ','); i >= 0 {
				v = v[:i]
			}
			v = strings.TrimSpace(v)
			if host, _, err := net.SplitHostPort(v); err == nil {
				v = host
			}
			if v != "" {
				return v
			}
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func isUnsafeMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return false
	default:
		return true
	}
}

func originAllowed(r *http.Request, allowed string) bool {
	origin := r.Header.Get("Origin")
	// Same-origin requests (server-to-server, some browsers) may omit Origin.
	return origin == "" || origin == allowed
}
