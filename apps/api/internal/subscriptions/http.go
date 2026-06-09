package subscriptions

import (
	"errors"
	"net/http"

	"github.com/lumni/mirante/internal/platform/respond"
)

const maxBody = 16 << 10

type handlers struct{ svc *Service }

// RegisterRoutes mounts the subscription routes, each wrapped with `protect`
// (session auth + CSRF). The composition root passes protect; this package never
// imports the auth package.
func RegisterRoutes(mux *http.ServeMux, protect func(http.Handler) http.Handler, svc *Service) {
	h := &handlers{svc: svc}
	mux.Handle("GET /api/subscriptions", protect(http.HandlerFunc(h.list)))
	mux.Handle("POST /api/subscriptions", protect(http.HandlerFunc(h.create)))
	mux.Handle("GET /api/subscriptions/{id}", protect(http.HandlerFunc(h.get)))
	mux.Handle("PATCH /api/subscriptions/{id}", protect(http.HandlerFunc(h.update)))
	mux.Handle("DELETE /api/subscriptions/{id}", protect(http.HandlerFunc(h.remove)))
}

func (h *handlers) list(w http.ResponseWriter, r *http.Request) {
	subs, err := h.svc.List(r.Context(), ListFilter{ProjectID: r.URL.Query().Get("project")})
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, map[string]any{"subscriptions": subs})
}

func (h *handlers) create(w http.ResponseWriter, r *http.Request) {
	var in CreateInput
	if err := respond.Decode(w, r, &in, maxBody); err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	sub, err := h.svc.Create(r.Context(), in)
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.JSON(w, http.StatusCreated, sub)
}

func (h *handlers) get(w http.ResponseWriter, r *http.Request) {
	sub, err := h.svc.Get(r.Context(), ID(r.PathValue("id")))
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, sub)
}

func (h *handlers) update(w http.ResponseWriter, r *http.Request) {
	var in UpdateInput
	if err := respond.Decode(w, r, &in, maxBody); err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	sub, err := h.svc.Update(r.Context(), ID(r.PathValue("id")), in)
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, sub)
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
		respond.Error(w, http.StatusNotFound, "not_found", "subscription not found")
	case errors.Is(err, ErrInvalid):
		respond.Error(w, http.StatusBadRequest, "validation_error", err.Error())
	default:
		respond.Error(w, http.StatusInternalServerError, "internal", "internal error")
	}
}
