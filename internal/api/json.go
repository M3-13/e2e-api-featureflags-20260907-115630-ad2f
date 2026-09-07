package api

import (
	"encoding/json"
	"net/http"
)

// writeJSON serializes v as JSON with a Content-Type of application/json and
// the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError writes the uniform JSON error object {"error": msg} with the
// given status code. msg must be generic and free of internal details.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
