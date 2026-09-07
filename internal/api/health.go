package api

import "net/http"

// healthHandler answers GET /healthz with 200 {"status":"ok"}.
func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
