package tasks

import (
	"errors"
	"net/http"

	"github.com/lumni/mirante/internal/platform/respond"
)

const maxBody = 16 << 10

type handlers struct{ svc *Service }

// RegisterRoutes mounts the task routes, each wrapped with `protect` (session
// auth + CSRF). The composition root passes protect; this package never imports
// the auth package.
func RegisterRoutes(mux *http.ServeMux, protect func(http.Handler) http.Handler, svc *Service) {
	h := &handlers{svc: svc}
	mux.Handle("GET /api/tasks", protect(http.HandlerFunc(h.list)))
	mux.Handle("POST /api/tasks", protect(http.HandlerFunc(h.create)))
	mux.Handle("GET /api/tasks/{id}", protect(http.HandlerFunc(h.get)))
	mux.Handle("PATCH /api/tasks/{id}", protect(http.HandlerFunc(h.update)))
	mux.Handle("DELETE /api/tasks/{id}", protect(http.HandlerFunc(h.remove)))
}

func (h *handlers) list(w http.ResponseWriter, r *http.Request) {
	ts, err := h.svc.List(r.Context(), ListFilter{
		Status:    r.URL.Query().Get("status"),
		ProjectID: r.URL.Query().Get("project"),
	})
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, map[string]any{"tasks": ts})
}

func (h *handlers) create(w http.ResponseWriter, r *http.Request) {
	var in CreateInput
	if err := respond.Decode(w, r, &in, maxBody); err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	t, err := h.svc.Create(r.Context(), in)
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.JSON(w, http.StatusCreated, t)
}

func (h *handlers) get(w http.ResponseWriter, r *http.Request) {
	t, err := h.svc.Get(r.Context(), ID(r.PathValue("id")))
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, t)
}

func (h *handlers) update(w http.ResponseWriter, r *http.Request) {
	var in UpdateInput
	if err := respond.Decode(w, r, &in, maxBody); err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	t, err := h.svc.Update(r.Context(), ID(r.PathValue("id")), in)
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, t)
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
		respond.Error(w, http.StatusNotFound, "not_found", "task not found")
	case errors.Is(err, ErrInvalid):
		respond.Error(w, http.StatusBadRequest, "validation_error", err.Error())
	default:
		respond.Error(w, http.StatusInternalServerError, "internal", "internal error")
	}
}
