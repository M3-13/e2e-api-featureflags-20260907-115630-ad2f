package store

import (
	"errors"
	"sort"
	"sync"
)

// ErrAlreadyExists is returned by Create when a flag with the same key
// already exists in the store.
var ErrAlreadyExists = errors.New("flag already exists")

// Flag is a single feature flag. It carries no user identifiers and no
// evaluation results (see AC-16).
type Flag struct {
	Key            string `json:"key"`
	Enabled        bool   `json:"enabled"`
	Description    string `json:"description"`
	RolloutPercent int    `json:"rollout_percent"`
}

// Store is a thread-safe in-memory feature-flag store guarded by a
// sync.RWMutex.
type Store struct {
	mu    sync.RWMutex
	flags map[string]Flag
}

// New returns an empty Store.
func New() *Store {
	return &Store{flags: make(map[string]Flag)}
}

// Create adds a flag, returning ErrAlreadyExists if the key is taken.
func (s *Store) Create(f Flag) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.flags[f.Key]; ok {
		return ErrAlreadyExists
	}
	s.flags[f.Key] = f
	return nil
}

// List returns all flags sorted by key.
func (s *Store) List() []Flag {
	s.mu.RLock()
	defer s.mu.RUnlock()
	keys := make([]string, 0, len(s.flags))
	for k := range s.flags {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	result := make([]Flag, 0, len(keys))
	for _, k := range keys {
		result = append(result, s.flags[k])
	}
	return result
}

// Get returns the flag with the given key and whether it exists.
func (s *Store) Get(key string) (Flag, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	f, ok := s.flags[key]
	return f, ok
}

// Update changes only the supplied (non-nil) fields of the flag with the
// given key. It returns the updated flag and whether the key existed.
func (s *Store) Update(key string, enabled *bool, description *string, rolloutPercent *int) (Flag, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	f, ok := s.flags[key]
	if !ok {
		return Flag{}, false
	}
	if enabled != nil {
		f.Enabled = *enabled
	}
	if description != nil {
		f.Description = *description
	}
	if rolloutPercent != nil {
		f.RolloutPercent = *rolloutPercent
	}
	s.flags[key] = f
	return f, true
}

// Delete removes the flag with the given key and reports whether it existed.
func (s *Store) Delete(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.flags[key]; !ok {
		return false
	}
	delete(s.flags, key)
	return true
}
