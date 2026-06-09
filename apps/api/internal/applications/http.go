package applications

import (
	"errors"
	"net/http"

	"github.com/lumni/mirante/internal/platform/respond"
)

const maxBody = 16 << 10

type handlers struct{ svc *Service }

// RegisterRoutes mounts the application routes, each wrapped with `protect`
// (session auth + CSRF). The composition root passes protect.
func RegisterRoutes(mux *http.ServeMux, protect func(http.Handler) http.Handler, svc *Service) {
	h := &handlers{svc: svc}
	mux.Handle("GET /api/applications", protect(http.HandlerFunc(h.list)))
	mux.Handle("POST /api/applications", protect(http.HandlerFunc(h.create)))
	mux.Handle("GET /api/applications/{id}", protect(http.HandlerFunc(h.get)))
	mux.Handle("PATCH /api/applications/{id}", protect(http.HandlerFunc(h.update)))
	mux.Handle("DELETE /api/applications/{id}", protect(http.HandlerFunc(h.remove)))
}

func (h *handlers) list(w http.ResponseWriter, r *http.Request) {
	as, err := h.svc.List(r.Context(), ListFilter{Status: r.URL.Query().Get("status")})
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, map[string]any{"applications": as})
}

func (h *handlers) create(w http.ResponseWriter, r *http.Request) {
	var in CreateInput
	if err := respond.Decode(w, r, &in, maxBody); err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	a, err := h.svc.Create(r.Context(), in)
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.JSON(w, http.StatusCreated, a)
}

func (h *handlers) get(w http.ResponseWriter, r *http.Request) {
	a, err := h.svc.Get(r.Context(), ID(r.PathValue("id")))
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, a)
}

func (h *handlers) update(w http.ResponseWriter, r *http.Request) {
	var in UpdateInput
	if err := respond.Decode(w, r, &in, maxBody); err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	a, err := h.svc.Update(r.Context(), ID(r.PathValue("id")), in)
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, a)
}

func (h *handlers) remove(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.Delete(r.Context(), ID(r.PathValue("id"))); err != nil {
		writeErr(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func writeErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		respond.Error(w, http.StatusNotFound, "not_found", "application not found")
	case errors.Is(err, ErrInvalid):
		respond.Error(w, http.StatusBadRequest, "validation_error", err.Error())
	default:
		respond.Error(w, http.StatusInternalServerError, "internal", "internal error")
	}
}
