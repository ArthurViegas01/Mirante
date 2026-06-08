// Package respond holds the shared HTTP JSON helpers used by every handler.
package respond

import (
	"encoding/json"
	"net/http"
)

// JSON writes v as a JSON response with the given status.
func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}

type errorBody struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// Error writes a structured JSON error.
func Error(w http.ResponseWriter, status int, code, message string) {
	var b errorBody
	b.Error.Code = code
	b.Error.Message = message
	JSON(w, status, b)
}

// Decode reads a JSON body (size-capped, rejecting unknown fields) into dst.
func Decode(w http.ResponseWriter, r *http.Request, dst any, maxBytes int64) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}
