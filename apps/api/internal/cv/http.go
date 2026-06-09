package cv

import (
	"errors"
	"net/http"

	"github.com/lumni/mirante/internal/platform/respond"
)

const maxBody = 128 << 10 // a full CV (experiences/education) can be large

type handlers struct{ svc *Service }

// RegisterRoutes mounts the profile routes, each wrapped with `protect` (session
// auth + CSRF). The composition root passes protect; this package never imports
// the auth package.
func RegisterRoutes(mux *http.ServeMux, protect func(http.Handler) http.Handler, svc *Service) {
	h := &handlers{svc: svc}
	mux.Handle("GET /api/profile", protect(http.HandlerFunc(h.get)))
	mux.Handle("PUT /api/profile", protect(http.HandlerFunc(h.save)))
	mux.Handle("PUT /api/cv", protect(http.HandlerFunc(h.saveCV)))
	mux.Handle("POST /api/cv/import", protect(http.HandlerFunc(h.importCV)))
	mux.Handle("GET /api/cv/export", protect(http.HandlerFunc(h.export)))
	mux.Handle("POST /api/cv/adapt", protect(http.HandlerFunc(h.adapt)))
}

func (h *handlers) adapt(w http.ResponseWriter, r *http.Request) {
	var in AdaptInput
	if err := respond.Decode(w, r, &in, maxBody); err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	res, err := h.svc.Adapt(r.Context(), in)
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, res)
}

func (h *handlers) export(w http.ResponseWriter, r *http.Request) {
	data, contentType, filename, err := h.svc.Export(r.Context(), r.URL.Query().Get("format"))
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal", "falha ao gerar o documento")
		return
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func (h *handlers) importCV(w http.ResponseWriter, r *http.Request) {
	var in ImportInput
	if err := respond.Decode(w, r, &in, maxBody); err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	draft, err := h.svc.ImportDraft(r.Context(), in)
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, draft)
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

func (h *handlers) saveCV(w http.ResponseWriter, r *http.Request) {
	var in CVInput
	if err := respond.Decode(w, r, &in, maxBody); err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	p, err := h.svc.SaveCV(r.Context(), in)
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
	case errors.Is(err, ErrLLMUnavailable):
		respond.Error(w, http.StatusServiceUnavailable, "llm_unavailable", "LLM não configurado (defina a API key)")
	default:
		respond.Error(w, http.StatusInternalServerError, "internal", "internal error")
	}
}
