package api

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// captureLogs redirects the default slog logger into a buffer and returns it,
// so tests can inspect the output. The returned func restores the original.
func captureLogs(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	h := slog.NewTextHandler(&buf, nil)
	prev := slog.Default()
	slog.SetDefault(slog.New(h))
	t.Cleanup(func() { slog.SetDefault(prev) })
	return &buf
}

func TestLoggingCapturesMethodPathStatusDuration(t *testing.T) {
	buf := captureLogs(t)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("created"))
	})

	handler := Logging(next)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/flags?key=secret", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201", rec.Code)
	}

	out := buf.String()
	if !strings.Contains(out, "method=POST") {
		t.Fatalf("log output missing method: %q", out)
	}
	if !strings.Contains(out, "path=/flags") {
		t.Fatalf("log output missing path: %q", out)
	}
	if strings.Contains(out, "key=secret") || strings.Contains(out, "?") {
		t.Fatalf("log output must not include query string: %q", out)
	}
	if !strings.Contains(out, "status=201") {
		t.Fatalf("log output missing status: %q", out)
	}
	if !strings.Contains(out, "duration=") {
		t.Fatalf("log output missing duration: %q", out)
	}
}

func TestLoggingDefaultStatusOK(t *testing.T) {
	buf := captureLogs(t)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})

	handler := Logging(next)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	handler.ServeHTTP(rec, req)

	out := buf.String()
	if !strings.Contains(out, "status=200") {
		t.Fatalf("expected default status 200, got: %q", out)
	}
}

func TestLoggingMasksCRLFInMethod(t *testing.T) {
	buf := captureLogs(t)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := Logging(next)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/flags", nil)
	req.Method = "GE\r\nT"
	handler.ServeHTTP(rec, req)

	out := buf.String()
	lines := strings.Count(out, "\n")
	if lines != 1 {
		t.Fatalf("expected exactly one log line, got %d: %q", lines+1, out)
	}
	if !strings.Contains(out, `method=GE\r\nT`) {
		t.Fatalf("CR/LF in method not masked: %q", out)
	}
}

func TestLoggingMasksCRLFInPath(t *testing.T) {
	buf := captureLogs(t)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := Logging(next)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/flags", nil)
	req.URL.Path = "/flags/evil\r\nline"
	handler.ServeHTTP(rec, req)

	out := buf.String()
	lines := strings.Count(out, "\n")
	if lines != 1 {
		t.Fatalf("expected exactly one log line, got %d: %q", lines+1, out)
	}
	if !strings.Contains(out, `path=/flags/evil\r\nline`) {
		t.Fatalf("CR/LF in path not masked: %q", out)
	}
}
