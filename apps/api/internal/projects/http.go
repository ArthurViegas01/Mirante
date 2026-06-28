package projects

import (
	"errors"
	"net/http"

	"github.com/lumni/mirante/internal/platform/respond"
)

const maxBody = 16 << 10

type handlers struct{ svc *Service }

// RegisterRoutes mounts the project routes, each wrapped with `protect`
// (session auth + CSRF). The composition root passes protect; this package
// never imports the auth package.
func RegisterRoutes(mux *http.ServeMux, protect func(http.Handler) http.Handler, svc *Service) {
	h := &handlers{svc: svc}
	mux.Handle("GET /api/projects", protect(http.HandlerFunc(h.list)))
	mux.Handle("POST /api/projects", protect(http.HandlerFunc(h.create)))
	mux.Handle("POST /api/projects/import", protect(http.HandlerFunc(h.importDraft)))
	mux.Handle("GET /api/projects/{id}", protect(http.HandlerFunc(h.get)))
	mux.Handle("PATCH /api/projects/{id}", protect(http.HandlerFunc(h.update)))
	mux.Handle("DELETE /api/projects/{id}", protect(http.HandlerFunc(h.remove)))
	mux.Handle("POST /api/projects/{id}/links", protect(http.HandlerFunc(h.addLink)))
	mux.Handle("DELETE /api/projects/{id}/links/{linkId}", protect(http.HandlerFunc(h.removeLink)))
}

func (h *handlers) list(w http.ResponseWriter, r *http.Request) {
	ps, err := h.svc.List(r.Context(), ListFilter{Status: r.URL.Query().Get("status")})
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, map[string]any{"projects": ps})
}

func (h *handlers) create(w http.ResponseWriter, r *http.Request) {
	var in CreateInput
	if err := respond.Decode(w, r, &in, maxBody); err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	p, err := h.svc.Create(r.Context(), in)
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.JSON(w, http.StatusCreated, p)
}

// importDraft fetches a GitHub repo and returns an unsaved project draft for the
// UI to pre-fill the new-project form. Nothing is persisted.
func (h *handlers) importDraft(w http.ResponseWriter, r *http.Request) {
	var in struct {
		URL string `json:"url"`
	}
	if err := respond.Decode(w, r, &in, maxBody); err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	draft, err := h.svc.ImportDraft(r.Context(), in.URL)
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, draft)
}

func (h *handlers) get(w http.ResponseWriter, r *http.Request) {
	p, err := h.svc.Get(r.Context(), ID(r.PathValue("id")))
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, p)
}

func (h *handlers) update(w http.ResponseWriter, r *http.Request) {
	var in UpdateInput
	if err := respond.Decode(w, r, &in, maxBody); err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	p, err := h.svc.Update(r.Context(), ID(r.PathValue("id")), in)
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, p)
}

func (h *handlers) remove(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.Delete(r.Context(), ID(r.PathValue("id"))); err != nil {
		writeErr(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *handlers) addLink(w http.ResponseWriter, r *http.Request) {
	var in LinkInput
	if err := respond.Decode(w, r, &in, maxBody); err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	p, err := h.svc.AddLink(r.Context(), ID(r.PathValue("id")), in)
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.JSON(w, http.StatusCreated, p)
}

func (h *handlers) removeLink(w http.ResponseWriter, r *http.Request) {
	err := h.svc.RemoveLink(r.Context(), ID(r.PathValue("id")), ID(r.PathValue("linkId")))
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func writeErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		respond.Error(w, http.StatusNotFound, "not_found", "project not found")
	case errors.Is(err, ErrInvalid):
		respond.Error(w, http.StatusBadRequest, "validation_error", err.Error())
	case errors.Is(err, ErrImportUnavailable):
		respond.Error(w, http.StatusServiceUnavailable, "import_unavailable", "import do GitHub indisponível")
	case errors.Is(err, ErrImportFailed):
		respond.Error(w, http.StatusUnprocessableEntity, "import_failed", err.Error())
	default:
		respond.Error(w, http.StatusInternalServerError, "internal", "internal error")
	}
}
