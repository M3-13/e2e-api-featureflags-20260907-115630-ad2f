package api

import (
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// statusRecorder wraps an http.ResponseWriter to capture the status code that
// a handler writes, defaulting to http.StatusOK when the handler never calls
// WriteHeader before writing its body.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	return r.ResponseWriter.Write(b)
}

// maskControl replaces carriage returns and line feeds with a visible
// placeholder so a single request can never inject additional log lines.
func maskControl(s string) string {
	s = strings.ReplaceAll(s, "\r", `\r`)
	s = strings.ReplaceAll(s, "\n", `\n`)
	return s
}

// Logging wraps next and logs one line per request with the method, the path
// without its query string, the resulting status code and the duration. Only
// method and path are logged; query parameters and user identifiers are never
// recorded. CR/LF in the method or path are masked so one request never spans
// multiple log lines.
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rec := &statusRecorder{ResponseWriter: w}
		next.ServeHTTP(rec, r)

		status := rec.status
		if status == 0 {
			status = http.StatusOK
		}

		slog.Info("request",
			"method", maskControl(r.Method),
			"path", maskControl(r.URL.Path),
			"status", status,
			"duration", time.Since(start).String(),
		)
	})
}
