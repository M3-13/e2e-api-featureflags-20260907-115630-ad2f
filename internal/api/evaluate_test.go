package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"featureflags/internal/store"
)

func newEvalRouter() *http.ServeMux {
	s := store.New()
	_ = s.Create(store.Flag{Key: "feature", Enabled: true, RolloutPercent: 50})
	_ = s.Create(store.Flag{Key: "off", Enabled: false, RolloutPercent: 100})
	_ = s.Create(store.Flag{Key: "full", Enabled: true, RolloutPercent: 100})
	return NewRouter(s)
}

func TestEvaluateReturnsConsistentResult(t *testing.T) {
	mux := newEvalRouter()

	var first map[string]any
	for i := 0; i < 10; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/flags/feature/evaluate?user=alice", nil)
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		var body map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("unmarshal body: %v", err)
		}
		if i == 0 {
			first = body
			continue
		}
		if body["key"] != first["key"] || body["result"] != first["result"] {
			t.Fatalf("call %d: %v != %v", i, body, first)
		}
	}
	if first["key"] != "feature" {
		t.Fatalf("key = %v, want feature", first["key"])
	}
	if _, ok := first["result"].(bool); !ok {
		t.Fatalf("result = %v (%T), want bool", first["result"], first["result"])
	}
}

func TestEvaluateMissingUserReturns400(t *testing.T) {
	mux := newEvalRouter()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/flags/feature/evaluate", nil)
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	if body["error"] == "" {
		t.Fatal("expected error message")
	}
}

func TestEvaluateUnknownKeyReturns404(t *testing.T) {
	mux := newEvalRouter()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/flags/nope/evaluate?user=alice", nil)
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestEvaluateResponseNeverContainsUser(t *testing.T) {
	mux := newEvalRouter()
	for _, u := range []string{"alice", "secret-user-42", "bob@example.com"} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/flags/feature/evaluate?user="+u, nil)
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("user %q: status = %d, want 200", u, rec.Code)
		}
		if strings.Contains(rec.Body.String(), u) {
			t.Fatalf("response leaked user value %q: %s", u, rec.Body.String())
		}
	}
}

func TestEvaluateDisabledFlagReturnsFalse(t *testing.T) {
	mux := newEvalRouter()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/flags/off/evaluate?user=alice", nil)
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	if body["result"] != false {
		t.Fatalf("result = %v, want false for disabled flag", body["result"])
	}
}

func TestEvaluateFullRolloutReturnsTrue(t *testing.T) {
	mux := newEvalRouter()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/flags/full/evaluate?user=alice", nil)
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	if body["result"] != true {
		t.Fatalf("result = %v, want true for 100%% rollout", body["result"])
	}
}
