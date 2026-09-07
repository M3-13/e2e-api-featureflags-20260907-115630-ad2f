package api

import (
	"net/http"

	"featureflags/internal/store"
)

// EvaluateHandler handles GET /flags/{key}/evaluate. Stub until ticket #2 fills it in.
func EvaluateHandler(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotImplemented, "not implemented")
	}
}
