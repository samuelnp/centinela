package audit

import (
	"strings"
	"testing"
)

// TestRatchetPartitions drives a real full-scan over a temp repo: one oversized
// file is baselined, a second is new, and a baseline-only entry is resolved.
func TestRatchetPartitions(t *testing.T) {
	cfg := tempRepo(t, "fail", map[string]string{
		"internal/keep.go": oversizedGo(0),
		"internal/new.go":  oversizedGo(5),
	})
	// Baseline captures only keep.go plus a phantom that no longer exists.
	keep := Compute("G1: File Size", []string{"internal/keep.go (110 lines)"})
	phantom := Compute("G1: File Size", []string{"internal/gone.go (200 lines)"})
	b := Baseline{Scheme: fingerprintScheme, Version: 1, Gates: []GateEntry{
		{Gate: "G1: File Size", Fingerprints: append(keep, phantom...)},
	}}

	d := Ratchet(cfg, b)
	if !d.HasNew() {
		t.Fatal("expected new violations")
	}
	if !containsRaw(d.New, "internal/new.go") {
		t.Fatalf("new.go not in New: %+v", d.New)
	}
	if !containsKey(d.Baselined, "internal/keep.go") {
		t.Fatalf("keep.go not Baselined: %+v", d.Baselined)
	}
	if !containsKey(d.Resolved, "internal/gone.go") {
		t.Fatalf("gone.go not Resolved: %+v", d.Resolved)
	}
}

// TestRatchetNoNewWhenAllBaselined matches the live violation in the baseline by
// its stable path key even though the recorded line count differs (AC-5).
func TestRatchetNoNewWhenAllBaselined(t *testing.T) {
	cfg := tempRepo(t, "fail", map[string]string{
		"internal/keep.go": oversizedGo(20),
	})
	// Baseline recorded at a different (smaller) line count — same key.
	fps := Compute("G1: File Size", []string{"internal/keep.go (101 lines)"})
	b := Baseline{Scheme: fingerprintScheme, Version: 1, Gates: []GateEntry{
		{Gate: "G1: File Size", Fingerprints: fps},
	}}
	d := Ratchet(cfg, b)
	if d.HasNew() {
		t.Fatalf("growth should not be new: %+v", d.New)
	}
	if len(d.Baselined) != 1 {
		t.Fatalf("want 1 baselined, got %d", len(d.Baselined))
	}
}

func containsRaw(fps []Fingerprint, sub string) bool {
	for _, fp := range fps {
		if strings.Contains(fp.Raw, sub) {
			return true
		}
	}
	return false
}

func containsKey(fps []Fingerprint, key string) bool {
	for _, fp := range fps {
		if fp.Key == key {
			return true
		}
	}
	return false
}
