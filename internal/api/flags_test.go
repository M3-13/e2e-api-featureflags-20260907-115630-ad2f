package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"featureflags/internal/store"
)

func newTestRouter() http.Handler {
	return NewRouter(store.New())
}

func doRequest(t *testing.T, h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	var reader *strings.Reader
	if body == "" {
		reader = strings.NewReader("")
	} else {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reader)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	return rr
}

func decodeFlag(t *testing.T, rr *httptest.ResponseRecorder) store.Flag {
	t.Helper()
	var f store.Flag
	if err := json.NewDecoder(rr.Body).Decode(&f); err != nil {
		t.Fatalf("failed to decode flag from response %q: %v", rr.Body.String(), err)
	}
	return f
}

func TestCreateFlagHandler(t *testing.T) {
	h := newTestRouter()

	rr := doRequest(t, h, http.MethodPost, "/flags", `{"key":"feature-a","enabled":true,"description":"desc","rollout_percent":50}`)
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d (body %q)", rr.Code, rr.Body.String())
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %q", ct)
	}
	flag := decodeFlag(t, rr)
	if flag.Key != "feature-a" || !flag.Enabled || flag.Description != "desc" || flag.RolloutPercent != 50 {
		t.Fatalf("unexpected flag: %+v", flag)
	}
}

func TestCreateFlagDefaults(t *testing.T) {
	h := newTestRouter()

	rr := doRequest(t, h, http.MethodPost, "/flags", `{"key":"feature-b","enabled":false}`)
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d (body %q)", rr.Code, rr.Body.String())
	}
	flag := decodeFlag(t, rr)
	if flag.Description != "" {
		t.Fatalf("expected empty description, got %q", flag.Description)
	}
	if flag.RolloutPercent != 0 {
		t.Fatalf("expected rollout_percent 0, got %d", flag.RolloutPercent)
	}
}

func TestCreateFlagDuplicate(t *testing.T) {
	h := newTestRouter()

	first := doRequest(t, h, http.MethodPost, "/flags", `{"key":"dup","enabled":true}`)
	if first.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", first.Code)
	}

	second := doRequest(t, h, http.MethodPost, "/flags", `{"key":"dup","enabled":false}`)
	if second.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d (body %q)", second.Code, second.Body.String())
	}
}

func TestCreateFlagInvalidJSON(t *testing.T) {
	h := newTestRouter()

	rr := doRequest(t, h, http.MethodPost, "/flags", `{not-json`)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d (body %q)", rr.Code, rr.Body.String())
	}
}

func TestCreateFlagMissingFields(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{"missing key", `{"enabled":true}`},
		{"empty key", `{"key":"","enabled":true}`},
		{"missing enabled", `{"key":"feature-x"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := newTestRouter()
			rr := doRequest(t, h, http.MethodPost, "/flags", tc.body)
			if rr.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d (body %q)", rr.Code, rr.Body.String())
			}
		})
	}
}

func TestCreateFlagRolloutOutOfRange(t *testing.T) {
	for _, pct := range []int{-1, 101} {
		h := newTestRouter()
		body := fmt.Sprintf(`{"key":"feature-r","enabled":true,"rollout_percent":%d}`, pct)
		rr := doRequest(t, h, http.MethodPost, "/flags", body)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 for rollout_percent %d, got %d (body %q)", pct, rr.Code, rr.Body.String())
		}
	}
}

func TestCreateFlagBodyTooLarge(t *testing.T) {
	h := newTestRouter()

	big := `{"key":"x","enabled":true,"description":"` + strings.Repeat("a", maxBodyBytes) + `"}`
	rr := doRequest(t, h, http.MethodPost, "/flags", big)
	if rr.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %d", rr.Code)
	}
}

func TestListFlagsEmpty(t *testing.T) {
	h := newTestRouter()

	rr := doRequest(t, h, http.MethodGet, "/flags", "")
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %q", ct)
	}
	if strings.TrimSpace(rr.Body.String()) != "[]" {
		t.Fatalf("expected empty JSON array, got %q", rr.Body.String())
	}
}

func TestListFlagsFilled(t *testing.T) {
	h := newTestRouter()

	doRequest(t, h, http.MethodPost, "/flags", `{"key":"b-flag","enabled":true}`)
	doRequest(t, h, http.MethodPost, "/flags", `{"key":"a-flag","enabled":false}`)

	rr := doRequest(t, h, http.MethodGet, "/flags", "")
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var flags []store.Flag
	if err := json.NewDecoder(rr.Body).Decode(&flags); err != nil {
		t.Fatalf("failed to decode list: %v", err)
	}
	if len(flags) != 2 {
		t.Fatalf("expected 2 flags, got %d", len(flags))
	}
	if flags[0].Key != "a-flag" || flags[1].Key != "b-flag" {
		t.Fatalf("expected sorted by key, got %+v", flags)
	}
}
