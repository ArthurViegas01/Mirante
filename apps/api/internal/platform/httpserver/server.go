// Package httpserver provides the HTTP middleware stack, the auth routes, and
// composition helpers. Domain modules register their own routes from the
// composition root (cmd/server); httpserver never imports a domain package.
package httpserver

import (
	"net/http"

	"github.com/lumni/mirante/internal/platform/respond"
)

// Healthz is the liveness endpoint.
func Healthz(w http.ResponseWriter, _ *http.Request) {
	respond.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
