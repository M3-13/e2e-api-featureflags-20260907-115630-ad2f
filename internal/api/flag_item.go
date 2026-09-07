package api

import (
	"net/http"

	"featureflags/internal/store"
)

// GetFlagHandler handles GET /flags/{key}. Stub until ticket #1 fills it in.
func GetFlagHandler(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotImplemented, "not implemented")
	}
}

// UpdateFlagHandler handles PUT /flags/{key}. Stub until ticket #1 fills it in.
func UpdateFlagHandler(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotImplemented, "not implemented")
	}
}

// DeleteFlagHandler handles DELETE /flags/{key}. Stub until ticket #1 fills it in.
func DeleteFlagHandler(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotImplemented, "not implemented")
	}
}
