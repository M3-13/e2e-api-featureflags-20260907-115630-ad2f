package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"featureflags/internal/store"
)

// maxBodyBytes bounds the request body size for handlers that read a JSON
// body (AC-11). 1 MiB.
const maxBodyBytes = 1 << 20

// GetFlagHandler handles GET /flags/{key}. It answers 200 with the flag, or
// 404 with the uniform JSON error object when the key is unknown.
func GetFlagHandler(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.PathValue("key")
		f, ok := s.Get(key)
		if !ok {
			writeError(w, http.StatusNotFound, "flag not found")
			return
		}
		writeJSON(w, http.StatusOK, f)
	}
}

// updateFlagRequest is the subset of fields a PUT /flags/{key} may carry.
// Pointer fields let us distinguish "absent" from "zero value".
type updateFlagRequest struct {
	Enabled        *bool   `json:"enabled"`
	Description    *string `json:"description"`
	RolloutPercent *int    `json:"rollout_percent"`
}

// UpdateFlagHandler handles PUT /flags/{key}. It updates only the supplied
// fields and answers 200 with the updated flag, 400 for an invalid body, 404
// for an unknown key, and 413 when the body exceeds maxBodyBytes.
func UpdateFlagHandler(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.PathValue("key")

		r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
		var req updateFlagRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			var maxErr *http.MaxBytesError
			if errors.As(err, &maxErr) {
				writeError(w, http.StatusRequestEntityTooLarge, "request body too large")
				return
			}
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if req.Enabled == nil && req.Description == nil && req.RolloutPercent == nil {
			writeError(w, http.StatusBadRequest, "at least one field is required")
			return
		}

		if req.RolloutPercent != nil && (*req.RolloutPercent < 0 || *req.RolloutPercent > 100) {
			writeError(w, http.StatusBadRequest, "rollout_percent must be between 0 and 100")
			return
		}

		updated, ok := s.Update(key, req.Enabled, req.Description, req.RolloutPercent)
		if !ok {
			writeError(w, http.StatusNotFound, "flag not found")
			return
		}

		writeJSON(w, http.StatusOK, updated)
	}
}

// DeleteFlagHandler handles DELETE /flags/{key}. It answers 204 with no body
// when the flag existed, or 404 when the key is unknown.
func DeleteFlagHandler(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.PathValue("key")
		if !s.Delete(key) {
			writeError(w, http.StatusNotFound, "flag not found")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
