package main

import (
	"os"
	"path/filepath"
	"testing"
)

// knownFeatureDir chdirs into a temp repo where requireKnownFeature(feature)
// passes (a .workflow/<feature>.json marker exists).
func knownFeatureDir(t *testing.T, feature string) string {
	t.Helper()
	d := t.TempDir()
	if err := os.MkdirAll(filepath.Join(d, ".workflow"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(d, ".workflow", feature+".json"),
		[]byte(`{"feature":"`+feature+`","currentStep":"plan","stepOrder":["plan"],"steps":{}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(d)
	return d
}

// TestCov2EvidenceInitLockError: the .lock sibling is present as a directory so
// evidence.Lock's OpenFile fails after requireKnownFeature passes.
func TestCov2EvidenceInitLockError(t *testing.T) {
	d := knownFeatureDir(t, "feat")
	if err := os.Mkdir(filepath.Join(d, ".workflow", "feat-gatekeeper.lock"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := runEvidenceInit(nil, []string{"feat", "gatekeeper"}); err == nil {
		t.Fatal("expected an evidence lock error")
	}
}

// TestCov2EvidenceInitWriteAtomicError: the target evidence JSON is present as a
// directory, so the atomic rename inside WriteAtomic fails.
func TestCov2EvidenceInitWriteAtomicError(t *testing.T) {
	d := knownFeatureDir(t, "feat")
	if err := os.Mkdir(filepath.Join(d, ".workflow", "feat-gatekeeper.json"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := runEvidenceInit(nil, []string{"feat", "gatekeeper"}); err == nil {
		t.Fatal("expected a WriteAtomic rename error")
	}
}

// TestCov2EvidenceInitCompanionError: the JSON write succeeds, but the companion
// .md target is a directory, so WriteCompanion's rename fails.
func TestCov2EvidenceInitCompanionError(t *testing.T) {
	d := knownFeatureDir(t, "feat")
	if err := os.Mkdir(filepath.Join(d, ".workflow", "feat-gatekeeper.md"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := runEvidenceInit(nil, []string{"feat", "gatekeeper"}); err == nil {
		t.Fatal("expected a WriteCompanion rename error")
	}
}
