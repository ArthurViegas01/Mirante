package httpserver

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/lumni/mirante/internal/platform/auth"
	"github.com/lumni/mirante/internal/platform/validate"
)

const maxLoginBody = 4 << 10 // 4 KiB

// AuthConfig configures cookie behavior and origin checks.
type AuthConfig struct {
	CookieName    string
	Secure        bool
	AllowedOrigin string
}

// AuthHandlers serves the auth routes and exposes auth middleware.
type AuthHandlers struct {
	svc *auth.Service
	cfg AuthConfig
}

// NewAuthHandlers builds the auth handler set.
func NewAuthHandlers(svc *auth.Service, cfg AuthConfig) *AuthHandlers {
	return &AuthHandlers{svc: svc, cfg: cfg}
}

// RegisterRoutes mounts the auth routes on the given mux.
func (h *AuthHandlers) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("GET /api/auth/status", http.HandlerFunc(h.Status))
	mux.Handle("POST /api/auth/signup", http.HandlerFunc(h.Signup))
	mux.Handle("POST /api/auth/login", http.HandlerFunc(h.Login))
	mux.Handle("GET /api/auth/me", h.RequireAuth(http.HandlerFunc(h.Me)))
	mux.Handle("POST /api/auth/logout", h.Protect(http.HandlerFunc(h.Logout)))
}

// Protect wraps a handler with session auth and CSRF enforcement. CSRF only
// applies to unsafe methods, so GET routes pass through unaffected. Domain
// modules wrap their whole sub-router with this.
func (h *AuthHandlers) Protect(next http.Handler) http.Handler {
	return h.RequireAuth(h.CSRF(next))
}

type loginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=1"`
}

type signupRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Name     string `json:"name" validate:"omitempty,max=80"`
}

// Status reports whether the instance still needs its first-run owner setup. It
// is public so the SPA can route an anonymous visitor to signup vs login.
func (h *AuthHandlers) Status(w http.ResponseWriter, r *http.Request) {
	needs, err := h.svc.NeedsSetup(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "internal error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"needs_setup": needs})
}

// Signup claims the instance for the first (and only) owner, then logs them in.
func (h *AuthHandlers) Signup(w http.ResponseWriter, r *http.Request) {
	if !originAllowed(r, h.cfg.AllowedOrigin) {
		writeError(w, http.StatusForbidden, "forbidden_origin", "origin not allowed")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxLoginBody)
	var req signupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	if err := validate.Struct(req); err != nil {
		writeError(w, http.StatusBadRequest, "validation_error",
			"a valid email and a password of at least 8 characters are required")
		return
	}

	sess, token, err := h.svc.Signup(r.Context(), req.Email, req.Password, req.Name, r.UserAgent(), clientIP(r))
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrSignupClosed):
			writeError(w, http.StatusForbidden, "signup_closed", "registration is closed")
		case errors.Is(err, auth.ErrInvalidCredentials):
			writeError(w, http.StatusBadRequest, "validation_error", "email and password are required")
		default:
			writeError(w, http.StatusInternalServerError, "internal", "internal error")
		}
		return
	}

	http.SetCookie(w, h.sessionCookie(token, sess.ExpiresAt))
	writeJSON(w, http.StatusCreated, map[string]any{"csrf_token": sess.CSRFToken})
}

// Login authenticates and starts a session.
func (h *AuthHandlers) Login(w http.ResponseWriter, r *http.Request) {
	if !originAllowed(r, h.cfg.AllowedOrigin) {
		writeError(w, http.StatusForbidden, "forbidden_origin", "origin not allowed")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxLoginBody)
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	if err := validate.Struct(req); err != nil {
		writeError(w, http.StatusBadRequest, "validation_error", "email and password are required")
		return
	}

	sess, token, err := h.svc.Login(r.Context(), req.Email, req.Password, r.UserAgent(), clientIP(r))
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrRateLimited):
			writeError(w, http.StatusTooManyRequests, "rate_limited", "too many attempts, try again later")
		case errors.Is(err, auth.ErrInvalidCredentials):
			writeError(w, http.StatusUnauthorized, "invalid_credentials", "invalid email or password")
		default:
			writeError(w, http.StatusInternalServerError, "internal", "internal error")
		}
		return
	}

	http.SetCookie(w, h.sessionCookie(token, sess.ExpiresAt))
	writeJSON(w, http.StatusOK, map[string]any{"csrf_token": sess.CSRFToken})
}

// Logout revokes the session and clears the cookie.
func (h *AuthHandlers) Logout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(h.cfg.CookieName); err == nil {
		_ = h.svc.Logout(r.Context(), c.Value)
	}
	http.SetCookie(w, h.clearCookie())
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Me returns the authenticated owner.
func (h *AuthHandlers) Me(w http.ResponseWriter, r *http.Request) {
	u, ok := UserFrom(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated", "login required")
		return
	}
	csrf := ""
	if sess, ok := SessionFrom(r.Context()); ok {
		csrf = sess.CSRFToken
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"user":       map[string]string{"id": u.ID, "email": u.Email, "name": u.Name},
		"csrf_token": csrf,
	})
}

// RequireAuth rejects requests without a valid session and injects the owner
// and session into the context.
func (h *AuthHandlers) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie(h.cfg.CookieName)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "unauthenticated", "login required")
			return
		}
		u, sess, err := h.svc.Authenticate(r.Context(), c.Value)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "unauthenticated", "login required")
			return
		}
		next.ServeHTTP(w, r.WithContext(withAuth(r.Context(), u, sess)))
	})
}

// CSRF enforces an Origin check and a matching X-CSRF-Token on unsafe methods.
// It must run inside RequireAuth (it reads the session from the context).
func (h *AuthHandlers) CSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isUnsafeMethod(r.Method) {
			if !originAllowed(r, h.cfg.AllowedOrigin) {
				writeError(w, http.StatusForbidden, "forbidden_origin", "origin not allowed")
				return
			}
			sess, ok := SessionFrom(r.Context())
			if !ok || subtle.ConstantTimeCompare(
				[]byte(r.Header.Get("X-CSRF-Token")), []byte(sess.CSRFToken)) != 1 {
				writeError(w, http.StatusForbidden, "csrf", "missing or invalid CSRF token")
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func (h *AuthHandlers) sessionCookie(token string, exp time.Time) *http.Cookie {
	return &http.Cookie{
		Name:     h.cfg.CookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   h.cfg.Secure,
		SameSite: http.SameSiteStrictMode,
		Expires:  exp,
		MaxAge:   int(time.Until(exp).Seconds()),
	}
}

func (h *AuthHandlers) clearCookie() *http.Cookie {
	return &http.Cookie{
		Name:     h.cfg.CookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   h.cfg.Secure,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	}
}
