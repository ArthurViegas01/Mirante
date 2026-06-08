package monitor

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/lumni/mirante/internal/platform/respond"
)

const maxBody = 16 << 10

type handlers struct{ mgr *Manager }

// RegisterRoutes mounts the monitor REST routes (the SSE stream endpoint is
// registered separately in the composition root). Each route is wrapped with
// `protect` (session auth + CSRF).
func RegisterRoutes(mux *http.ServeMux, protect func(http.Handler) http.Handler, mgr *Manager) {
	h := &handlers{mgr: mgr}
	mux.Handle("GET /api/services", protect(http.HandlerFunc(h.listServices)))
	mux.Handle("POST /api/services", protect(http.HandlerFunc(h.createService)))
	mux.Handle("GET /api/services/{id}", protect(http.HandlerFunc(h.serviceDetail)))
	mux.Handle("PATCH /api/services/{id}", protect(http.HandlerFunc(h.updateService)))
	mux.Handle("DELETE /api/services/{id}", protect(http.HandlerFunc(h.deleteService)))
	mux.Handle("POST /api/services/{id}/enabled", protect(http.HandlerFunc(h.setEnabled)))

	mux.Handle("GET /api/alerts", protect(http.HandlerFunc(h.listAlerts)))
	mux.Handle("POST /api/alerts/{id}/read", protect(http.HandlerFunc(h.markRead)))
	mux.Handle("POST /api/alerts/read-all", protect(http.HandlerFunc(h.markAllRead)))
}

func (h *handlers) listServices(w http.ResponseWriter, r *http.Request) {
	services, err := h.mgr.ListServices(r.Context(), r.URL.Query().Get("project_id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, map[string]any{"services": services})
}

func (h *handlers) createService(w http.ResponseWriter, r *http.Request) {
	var in CreateServiceInput
	if err := respond.Decode(w, r, &in, maxBody); err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	svc, err := h.mgr.CreateService(r.Context(), in)
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.JSON(w, http.StatusCreated, svc)
}

func (h *handlers) serviceDetail(w http.ResponseWriter, r *http.Request) {
	d, err := h.mgr.Detail(r.Context(), ServiceID(r.PathValue("id")))
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, d)
}

func (h *handlers) updateService(w http.ResponseWriter, r *http.Request) {
	var in UpdateServiceInput
	if err := respond.Decode(w, r, &in, maxBody); err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	svc, err := h.mgr.UpdateService(r.Context(), ServiceID(r.PathValue("id")), in)
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, svc)
}

func (h *handlers) deleteService(w http.ResponseWriter, r *http.Request) {
	if err := h.mgr.DeleteService(r.Context(), ServiceID(r.PathValue("id"))); err != nil {
		writeErr(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *handlers) setEnabled(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Enabled bool `json:"enabled"`
	}
	if err := respond.Decode(w, r, &in, 1<<10); err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	svc, err := h.mgr.SetEnabled(r.Context(), ServiceID(r.PathValue("id")), in.Enabled)
	if err != nil {
		writeErr(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, svc)
}

func (h *handlers) listAlerts(w http.ResponseWriter, r *http.Request) {
	unread := r.URL.Query().Get("unread") == "1"
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	alerts, err := h.mgr.ListAlerts(r.Context(), limit, unread)
	if err != nil {
		writeErr(w, err)
		return
	}
	count, _ := h.mgr.UnreadCount(r.Context())
	respond.JSON(w, http.StatusOK, map[string]any{"alerts": alerts, "unread_count": count})
}

func (h *handlers) markRead(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", "invalid id")
		return
	}
	if err := h.mgr.MarkAlertRead(r.Context(), id); err != nil {
		writeErr(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *handlers) markAllRead(w http.ResponseWriter, r *http.Request) {
	if err := h.mgr.MarkAllAlertsRead(r.Context()); err != nil {
		writeErr(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func writeErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		respond.Error(w, http.StatusNotFound, "not_found", "service not found")
	case errors.Is(err, ErrInvalid):
		respond.Error(w, http.StatusBadRequest, "validation_error", err.Error())
	default:
		respond.Error(w, http.StatusInternalServerError, "internal", "internal error")
	}
}
