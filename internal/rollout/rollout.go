// Package rollout implements the deterministic rollout decision for feature
// flags. The decision is a pure function of (key, user) with no time or
// randomness, so the same inputs always produce the same result across
// process restarts.
package rollout

import "hash/fnv"

// Evaluate decides whether a flag is on for the given user.
//
// The result is deterministic:
//   - enabled == false        -> false
//   - rolloutPercent <= 0     -> false
//   - rolloutPercent >= 100   -> true
//   - otherwise a stable FNV-1a hash of key+":"+user modulo 100 is compared
//     against rolloutPercent.
func Evaluate(key, user string, rolloutPercent int, enabled bool) bool {
	if !enabled {
		return false
	}
	if rolloutPercent <= 0 {
		return false
	}
	if rolloutPercent >= 100 {
		return true
	}

	h := fnv.New32a()
	_, _ = h.Write([]byte(key + ":" + user))
	bucket := int(h.Sum32() % 100)
	return bucket < rolloutPercent
}
