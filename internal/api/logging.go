package api

import "net/http"

// Logging wraps next in the access-logging middleware. Currently a pure
// pass-through; ticket #5 implements the real logging behaviour.
func Logging(next http.Handler) http.Handler {
	return next
}
