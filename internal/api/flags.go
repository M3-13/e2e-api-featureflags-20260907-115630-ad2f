package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"featureflags/internal/store"
)

// maxBodyBytes is the maximum accepted request body size for POST /flags.
const maxBodyBytes = 1 << 20 // 1 MiB

// createFlagRequest is the decoded JSON body of POST /flags.
// Enabled and RolloutPercent are pointers so that a missing field can be
// told apart from a zero value.
type createFlagRequest struct {
	Key            string `json:"key"`
	Enabled        *bool  `json:"enabled"`
	Description    string `json:"description"`
	RolloutPercent *int   `json:"rollout_percent"`
}

// CreateFlagHandler handles POST /flags.
func CreateFlagHandler(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

		var req createFlagRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			var maxErr *http.MaxBytesError
			if errors.As(err, &maxErr) {
				writeError(w, http.StatusRequestEntityTooLarge, "request body too large")
				return
			}
			writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}

		if req.Key == "" {
			writeError(w, http.StatusBadRequest, "key is required")
			return
		}
		if req.Enabled == nil {
			writeError(w, http.StatusBadRequest, "enabled is required")
			return
		}

		rolloutPercent := 0
		if req.RolloutPercent != nil {
			if *req.RolloutPercent < 0 || *req.RolloutPercent > 100 {
				writeError(w, http.StatusBadRequest, "rollout_percent must be between 0 and 100")
				return
			}
			rolloutPercent = *req.RolloutPercent
		}

		flag := store.Flag{
			Key:            req.Key,
			Enabled:        *req.Enabled,
			Description:    req.Description,
			RolloutPercent: rolloutPercent,
		}

		if err := s.Create(flag); err != nil {
			if errors.Is(err, store.ErrAlreadyExists) {
				writeError(w, http.StatusConflict, "flag already exists")
				return
			}
			writeError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		writeJSON(w, http.StatusCreated, flag)
	}
}

// ListFlagsHandler handles GET /flags.
func ListFlagsHandler(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, s.List())
	}
}
