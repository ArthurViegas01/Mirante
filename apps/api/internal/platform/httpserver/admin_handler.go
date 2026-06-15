package httpserver

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/lumni/mirante/internal/platform/auth"
	"github.com/lumni/mirante/internal/platform/validate"
)

// adminUserView is the admin-facing user shape (never includes the password hash).
type adminUserView struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

func toAdminUserView(u *auth.User) adminUserView {
	return adminUserView{
		ID: u.ID, Email: u.Email, Name: u.Name, Role: u.Role, Status: u.Status, CreatedAt: u.CreatedAt,
	}
}

// AdminListUsers returns every account (admin only).
func (h *AuthHandlers) AdminListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.svc.ListUsers(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "internal error")
		return
	}
	out := make([]adminUserView, 0, len(users))
	for _, u := range users {
		out = append(out, toAdminUserView(u))
	}
	writeJSON(w, http.StatusOK, map[string]any{"users": out})
}

type adminCreateUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Name     string `json:"name" validate:"omitempty,max=80"`
	Role     string `json:"role" validate:"omitempty,oneof=admin user"`
}

// AdminCreateUser creates an already-active account directly (admin only).
func (h *AuthHandlers) AdminCreateUser(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxLoginBody)
	var req adminCreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	if err := validate.Struct(req); err != nil {
		writeError(w, http.StatusBadRequest, "validation_error",
			"a valid email and a password of at least 8 characters are required")
		return
	}

	u, err := h.svc.AdminCreateUser(r.Context(), req.Email, req.Password, req.Name, req.Role)
	switch {
	case err == nil:
		writeJSON(w, http.StatusCreated, toAdminUserView(u))
	case errors.Is(err, auth.ErrEmailTaken):
		writeError(w, http.StatusConflict, "email_taken", "this email is already registered")
	case errors.Is(err, auth.ErrInvalidCredentials):
		writeError(w, http.StatusBadRequest, "validation_error",
			"a valid email and a password of at least 8 characters are required")
	default:
		writeError(w, http.StatusInternalServerError, "internal", "internal error")
	}
}

// AdminActivateUser activates a pending/disabled account.
func (h *AuthHandlers) AdminActivateUser(w http.ResponseWriter, r *http.Request) {
	h.setUserStatus(w, r, true)
}

// AdminDeactivateUser disables an account (and drops its sessions).
func (h *AuthHandlers) AdminDeactivateUser(w http.ResponseWriter, r *http.Request) {
	h.setUserStatus(w, r, false)
}

func (h *AuthHandlers) setUserStatus(w http.ResponseWriter, r *http.Request, activate bool) {
	id := r.PathValue("id")
	if self, ok := UserFrom(r.Context()); ok && self.ID == id && !activate {
		writeError(w, http.StatusBadRequest, "self_action", "you cannot deactivate your own account")
		return
	}

	var err error
	if activate {
		err = h.svc.ActivateUser(r.Context(), id)
	} else {
		err = h.svc.DeactivateUser(r.Context(), id)
	}
	switch {
	case err == nil:
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	case errors.Is(err, auth.ErrUserNotFound):
		writeError(w, http.StatusNotFound, "not_found", "user not found")
	default:
		writeError(w, http.StatusInternalServerError, "internal", "internal error")
	}
}

// AdminDeleteUser removes an account and all of its data (admin only).
func (h *AuthHandlers) AdminDeleteUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if self, ok := UserFrom(r.Context()); ok && self.ID == id {
		writeError(w, http.StatusBadRequest, "self_action", "you cannot delete your own account")
		return
	}
	if err := h.svc.DeleteUser(r.Context(), id); err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal", "internal error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
