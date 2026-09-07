package api

import (
	"net/http"

	"featureflags/internal/rollout"
	"featureflags/internal/store"
)

// EvaluateHandler handles GET /flags/{key}/evaluate?user={id}.
//
// The user query parameter is required: a missing user answers 400. An
// unknown key answers 404. Otherwise the response is 200 with
// {"key": <flag-key>, "result": bool}. The response never contains the user
// value (AC-17).
func EvaluateHandler(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.PathValue("key")
		user := r.URL.Query().Get("user")
		if user == "" {
			writeError(w, http.StatusBadRequest, "missing user parameter")
			return
		}

		flag, ok := s.Get(key)
		if !ok {
			writeError(w, http.StatusNotFound, "flag not found")
			return
		}

		result := rollout.Evaluate(key, user, flag.RolloutPercent, flag.Enabled)
		writeJSON(w, http.StatusOK, map[string]any{
			"key":    key,
			"result": result,
		})
	}
}
