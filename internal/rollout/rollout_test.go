package rollout

import "testing"

func TestEvaluateDisabledAlwaysFalse(t *testing.T) {
	if Evaluate("key", "user", 50, false) {
		t.Fatal("enabled=false must return false regardless of rolloutPercent")
	}
	if Evaluate("key", "user", 100, false) {
		t.Fatal("enabled=false must return false even at rolloutPercent 100")
	}
}

func TestEvaluateRolloutBounds(t *testing.T) {
	if Evaluate("key", "user", 0, true) {
		t.Fatal("rolloutPercent 0 must return false")
	}
	if !Evaluate("key", "user", 100, true) {
		t.Fatal("rolloutPercent 100 must return true")
	}
	if Evaluate("key", "user", -5, true) {
		t.Fatal("negative rolloutPercent must return false")
	}
}

func TestEvaluateStableAcrossCalls(t *testing.T) {
	first := Evaluate("my-feature", "alice", 30, true)
	for i := 0; i < 1000; i++ {
		if got := Evaluate("my-feature", "alice", 30, true); got != first {
			t.Fatalf("result changed on call %d: %v != %v", i, got, first)
		}
	}
}

func TestEvaluateDifferentUsersDifferDeterministically(t *testing.T) {
	// The same (key, user, percent) must always agree; users must be able to
	// differ but each individually stable.
	got := map[string]bool{}
	for _, u := range []string{"alice", "bob", "carol", "dave", "eve"} {
		got[u] = Evaluate("feature", u, 50, true)
	}
	// Re-evaluate: identical to the first pass.
	for u, want := range got {
		if Evaluate("feature", u, 50, true) != want {
			t.Fatalf("user %q changed between passes", u)
		}
	}
}

func TestEvaluateDistributionReasonable(t *testing.T) {
	on := 0
	const n = 10000
	for i := 0; i < n; i++ {
		if Evaluate("feature", userN(i), 50, true) {
			on++
		}
	}
	// 50% rollout over 10000 users: expect roughly half, with generous slack.
	if on < 4000 || on > 6000 {
		t.Fatalf("rollout 50%% produced %d/%d on; expected ~5000", on, n)
	}
}

func userN(i int) string {
	const digits = "0123456789abcdefghijklmnopqrstuvwxyz"
	var b [8]byte
	for j := range b {
		b[j] = digits[i%len(digits)]
		i /= len(digits)
	}
	return string(b[:])
}
