package intake

import (
	"errors"
	"net/http"

	"github.com/lumni/mirante/internal/platform/respond"
)

type handlers struct{ svc *Service }

// RegisterRoutes mounts the intake (freelance funnel) routes, each wrapped with
// `protect` (session auth + CSRF). The composition root passes protect; this
// package never imports the auth layer.
func RegisterRoutes(mux *http.ServeMux, protect func(http.Handler) http.Handler, svc *Service) {
	h := &handlers{svc: svc}
	mux.Handle("GET /api/intake", protect(http.HandlerFunc(h.list)))
	mux.Handle("POST /api/intake/{id}/dismiss", protect(http.HandlerFunc(h.dismiss)))
}

// list returns staged opportunities, highest score first. Query params:
//
//	?estado=novo|descartado|promovido  — filter by lifecycle state (default: all)
//	?shortlist=true                    — keep only score >= the configured floor
func (h *handlers) list(w http.ResponseWriter, r *http.Request) {
	f := ListFilter{Estado: Estado(r.URL.Query().Get("estado"))}
	if r.URL.Query().Get("shortlist") == "true" {
		f.MinScore = h.svc.MinScore()
	}
	items, err := h.svc.List(r.Context(), f)
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, map[string]any{"items": items, "min_score": h.svc.MinScore()})
}

func (h *handlers) dismiss(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.Dismiss(r.Context(), ID(r.PathValue("id"))); err != nil {
		writeErr(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, map[string]string{"status": string(EstadoDescartado)})
}

func writeErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		respond.Error(w, http.StatusNotFound, "not_found", "item não encontrado")
	default:
		respond.Error(w, http.StatusInternalServerError, "internal", "internal error")
	}
}
