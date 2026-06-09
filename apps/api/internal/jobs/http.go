package jobs

import (
	"errors"
	"net/http"

	"github.com/lumni/mirante/internal/platform/respond"
)

const maxBody = 64 << 10 // job descriptions can be long

type handlers struct{ svc *Service }

// RegisterRoutes mounts the job routes, each wrapped with `protect` (session auth
// + CSRF). The composition root passes protect; this package never imports auth.
func RegisterRoutes(mux *http.ServeMux, protect func(http.Handler) http.Handler, svc *Service) {
	h := &handlers{svc: svc}
	mux.Handle("GET /api/jobs", protect(http.HandlerFunc(h.list)))
	mux.Handle("POST /api/jobs", protect(http.HandlerFunc(h.create)))
	mux.Handle("GET /api/jobs/{id}", protect(http.HandlerFunc(h.get)))
	mux.Handle("PATCH /api/jobs/{id}", protect(http.HandlerFunc(h.update)))
	mux.Handle("DELETE /api/jobs/{id}", protect(http.HandlerFunc(h.remove)))
	mux.Handle("POST /api/jobs/{id}/enrich", protect(http.HandlerFunc(h.enrich)))
	mux.Handle("POST /api/jobs/import", protect(http.HandlerFunc(h.importDraft)))
}

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

func (h *handlers) list(w http.ResponseWriter, r *http.Request) {
	js, err := h.svc.List(r.Context())
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, map[string]any{"jobs": js})
}

func (h *handlers) create(w http.ResponseWriter, r *http.Request) {
	var in CreateInput
	if err := respond.Decode(w, r, &in, maxBody); err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	j, err := h.svc.Create(r.Context(), in)
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.JSON(w, http.StatusCreated, j)
}

func (h *handlers) get(w http.ResponseWriter, r *http.Request) {
	j, err := h.svc.Get(r.Context(), ID(r.PathValue("id")))
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, j)
}

func (h *handlers) update(w http.ResponseWriter, r *http.Request) {
	var in UpdateInput
	if err := respond.Decode(w, r, &in, maxBody); err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	j, err := h.svc.Update(r.Context(), ID(r.PathValue("id")), in)
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, j)
}

func (h *handlers) remove(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.Delete(r.Context(), ID(r.PathValue("id"))); err != nil {
		writeErr(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *handlers) enrich(w http.ResponseWriter, r *http.Request) {
	j, err := h.svc.Enrich(r.Context(), ID(r.PathValue("id")))
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, j)
}

func writeErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		respond.Error(w, http.StatusNotFound, "not_found", "job not found")
	case errors.Is(err, ErrInvalid):
		respond.Error(w, http.StatusBadRequest, "validation_error", err.Error())
	case errors.Is(err, ErrLLMUnavailable):
		respond.Error(w, http.StatusServiceUnavailable, "llm_unavailable", "LLM não configurado (defina a API key)")
	case errors.Is(err, ErrImportUnavailable):
		respond.Error(w, http.StatusServiceUnavailable, "import_unavailable", "import de link indisponível")
	case errors.Is(err, ErrImportFailed):
		respond.Error(w, http.StatusUnprocessableEntity, "import_failed", "não consegui ler a vaga desse link — cole a descrição manualmente")
	default:
		respond.Error(w, http.StatusInternalServerError, "internal", "internal error")
	}
}
