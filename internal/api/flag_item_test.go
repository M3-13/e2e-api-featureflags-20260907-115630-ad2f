package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"featureflags/internal/store"
)

func seedFlag(t *testing.T, s *store.Store, f store.Flag) {
	t.Helper()
	if err := s.Create(f); err != nil {
		t.Fatalf("seed flag %q: %v", f.Key, err)
	}
}

func TestGetFlagExisting(t *testing.T) {
	s := store.New()
	seedFlag(t, s, store.Flag{Key: "alpha", Enabled: true, Description: "d", RolloutPercent: 42})

	mux := NewRouter(s)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/flags/alpha", nil)
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", ct)
	}

	var got store.Flag
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	if got.Key != "alpha" || !got.Enabled || got.Description != "d" || got.RolloutPercent != 42 {
		t.Fatalf("got %+v, want seeded flag", got)
	}
}

func TestGetFlagMissing(t *testing.T) {
	mux := NewRouter(store.New())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/flags/nope", nil)
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", ct)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	if body["error"] == "" {
		t.Fatal("expected error field in body")
	}
}

func TestUpdateFlagOnlySuppliedFields(t *testing.T) {
	s := store.New()
	seedFlag(t, s, store.Flag{Key: "beta", Enabled: false, Description: "orig", RolloutPercent: 10})

	mux := NewRouter(s)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/flags/beta",
		strings.NewReader(`{"description":"changed"}`))
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var got store.Flag
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	// Supplied field changed, un-supplied fields unchanged.
	if got.Description != "changed" {
		t.Fatalf("description = %q, want changed", got.Description)
	}
	if got.Enabled != false {
		t.Fatalf("enabled = %v, want false (unchanged)", got.Enabled)
	}
	if got.RolloutPercent != 10 {
		t.Fatalf("rollout_percent = %d, want 10 (unchanged)", got.RolloutPercent)
	}
	if got.Key != "beta" {
		t.Fatalf("key = %q, want beta (unchanged)", got.Key)
	}
}

func TestUpdateFlagNoFields(t *testing.T) {
	s := store.New()
	seedFlag(t, s, store.Flag{Key: "gamma", Enabled: true, RolloutPercent: 5})

	mux := NewRouter(s)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/flags/gamma",
		strings.NewReader(`{}`))
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestUpdateFlagUnknownKey(t *testing.T) {
	mux := NewRouter(store.New())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/flags/ghost",
		strings.NewReader(`{"enabled":true}`))
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestUpdateFlagInvalidJSON(t *testing.T) {
	s := store.New()
	seedFlag(t, s, store.Flag{Key: "delta", Enabled: true})

	mux := NewRouter(s)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/flags/delta",
		strings.NewReader(`{not json`))
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestUpdateFlagRolloutPercentOutOfRange(t *testing.T) {
	s := store.New()
	seedFlag(t, s, store.Flag{Key: "eps", Enabled: true})

	mux := NewRouter(s)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/flags/eps",
		strings.NewReader(`{"rollout_percent":101}`))
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestUpdateFlagTooLarge(t *testing.T) {
	s := store.New()
	seedFlag(t, s, store.Flag{Key: "big", Enabled: true})

	big := `{"description":"` + strings.Repeat("a", maxBodyBytes+1) + `"}`

	mux := NewRouter(s)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/flags/big",
		strings.NewReader(big))
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want 413", rec.Code)
	}
}

func TestDeleteFlagThenGet(t *testing.T) {
	s := store.New()
	seedFlag(t, s, store.Flag{Key: "zeta", Enabled: true})

	mux := NewRouter(s)

	delRec := httptest.NewRecorder()
	delReq := httptest.NewRequest(http.MethodDelete, "/flags/zeta", nil)
	mux.ServeHTTP(delRec, delReq)

	if delRec.Code != http.StatusNoContent {
		t.Fatalf("delete status = %d, want 204", delRec.Code)
	}
	if delRec.Body.Len() != 0 {
		t.Fatalf("delete body = %q, want empty", delRec.Body.String())
	}

	getRec := httptest.NewRecorder()
	getReq := httptest.NewRequest(http.MethodGet, "/flags/zeta", nil)
	mux.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusNotFound {
		t.Fatalf("get-after-delete status = %d, want 404", getRec.Code)
	}
}

func TestDeleteFlagUnknownKey(t *testing.T) {
	mux := NewRouter(store.New())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/flags/ghost", nil)
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}
