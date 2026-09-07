package api

import (
	"net/http"

	"featureflags/internal/store"
)

// CreateFlagHandler handles POST /flags. Stub until ticket #4 fills it in.
func CreateFlagHandler(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotImplemented, "not implemented")
	}
}

// ListFlagsHandler handles GET /flags. Stub until ticket #4 fills it in.
func ListFlagsHandler(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotImplemented, "not implemented")
	}
}
