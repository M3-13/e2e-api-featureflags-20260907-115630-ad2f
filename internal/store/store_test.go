package store

import (
	"errors"
	"testing"
)

func TestCreateAndGet(t *testing.T) {
	s := New()
	f := Flag{Key: "alpha", Enabled: true, Description: "d", RolloutPercent: 50}
	if err := s.Create(f); err != nil {
		t.Fatalf("Create: %v", err)
	}
	got, ok := s.Get("alpha")
	if !ok {
		t.Fatal("expected flag to exist")
	}
	if got != f {
		t.Fatalf("got %+v, want %+v", got, f)
	}
}

func TestCreateDuplicateReturnsErrAlreadyExists(t *testing.T) {
	s := New()
	if err := s.Create(Flag{Key: "a"}); err != nil {
		t.Fatalf("Create: %v", err)
	}
	err := s.Create(Flag{Key: "a"})
	if !errors.Is(err, ErrAlreadyExists) {
		t.Fatalf("want ErrAlreadyExists, got %v", err)
	}
}

func TestListSortedByKey(t *testing.T) {
	s := New()
	for _, k := range []string{"c", "a", "b"} {
		if err := s.Create(Flag{Key: k}); err != nil {
			t.Fatalf("Create(%q): %v", k, err)
		}
	}
	got := s.List()
	if len(got) != 3 {
		t.Fatalf("len = %d, want 3", len(got))
	}
	for i, want := range []string{"a", "b", "c"} {
		if got[i].Key != want {
			t.Fatalf("got[%d].Key = %q, want %q", i, got[i].Key, want)
		}
	}
}

func TestListEmpty(t *testing.T) {
	s := New()
	if got := s.List(); len(got) != 0 {
		t.Fatalf("want empty list, got %v", got)
	}
}

func TestGetMissing(t *testing.T) {
	s := New()
	if _, ok := s.Get("missing"); ok {
		t.Fatal("expected missing key to return false")
	}
}

func TestUpdateOnlySuppliedFields(t *testing.T) {
	s := New()
	if err := s.Create(Flag{Key: "a", Enabled: false, Description: "orig", RolloutPercent: 10}); err != nil {
		t.Fatalf("Create: %v", err)
	}
	enabled := true
	desc := "new"
	got, ok := s.Update("a", &enabled, &desc, nil)
	if !ok {
		t.Fatal("expected update to succeed")
	}
	if got.Enabled != true || got.Description != "new" || got.RolloutPercent != 10 {
		t.Fatalf("got %+v; enabled/description updated, rollout kept", got)
	}
}

func TestUpdateMissingKey(t *testing.T) {
	s := New()
	if _, ok := s.Update("nope", nil, nil, nil); ok {
		t.Fatal("expected update on missing key to return false")
	}
}

func TestDelete(t *testing.T) {
	s := New()
	if err := s.Create(Flag{Key: "a"}); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if !s.Delete("a") {
		t.Fatal("expected first Delete to return true")
	}
	if s.Delete("a") {
		t.Fatal("expected second Delete to return false")
	}
	if _, ok := s.Get("a"); ok {
		t.Fatal("flag should be gone after delete")
	}
}
