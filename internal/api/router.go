package api

import (
	"net/http"

	"featureflags/internal/store"
)

// NewRouter builds the application mux and registers every route.
func NewRouter(s *store.Store) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /flags", CreateFlagHandler(s))
	mux.HandleFunc("GET /flags", ListFlagsHandler(s))
	mux.HandleFunc("GET /flags/{key}", GetFlagHandler(s))
	mux.HandleFunc("PUT /flags/{key}", UpdateFlagHandler(s))
	mux.HandleFunc("DELETE /flags/{key}", DeleteFlagHandler(s))
	mux.HandleFunc("GET /flags/{key}/evaluate", EvaluateHandler(s))
	mux.HandleFunc("GET /healthz", healthHandler)
	return mux
}
