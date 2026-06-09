package cv

import (
	"errors"
	"net/http"

	"github.com/lumni/mirante/internal/platform/respond"
)

const maxBody = 16 << 10

type handlers struct{ svc *Service }

// RegisterRoutes mounts the profile routes, each wrapped with `protect` (session
// auth + CSRF). The composition root passes protect; this package never imports
// the auth package.
func RegisterRoutes(mux *http.ServeMux, protect func(http.Handler) http.Handler, svc *Service) {
	h := &handlers{svc: svc}
	mux.Handle("GET /api/profile", protect(http.HandlerFunc(h.get)))
	mux.Handle("PUT /api/profile", protect(http.HandlerFunc(h.save)))
}

func (h *handlers) get(w http.ResponseWriter, r *http.Request) {
	p, err := h.svc.GetProfile(r.Context())
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, p)
}

func (h *handlers) save(w http.ResponseWriter, r *http.Request) {
	var in ProfileInput
	if err := respond.Decode(w, r, &in, maxBody); err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	p, err := h.svc.SaveProfile(r.Context(), in)
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, p)
}

func writeErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalid):
		respond.Error(w, http.StatusBadRequest, "validation_error", err.Error())
	default:
		respond.Error(w, http.StatusInternalServerError, "internal", "internal error")
	}
}
