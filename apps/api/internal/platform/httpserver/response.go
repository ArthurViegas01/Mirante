package httpserver

import (
	"net/http"

	"github.com/lumni/mirante/internal/platform/respond"
)

// Thin internal aliases so existing handlers keep their call sites while the
// canonical helpers live in the shared respond package.
func writeJSON(w http.ResponseWriter, status int, v any) { respond.JSON(w, status, v) }

func writeError(w http.ResponseWriter, status int, code, message string) {
	respond.Error(w, status, code, message)
}
