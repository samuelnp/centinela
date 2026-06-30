package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// seedCorruptEvidence chdirs into a temp dir holding a malformed evidence JSON
// for (feature, role) so evidence.Read returns a non-NotFound parse error.
func seedCorruptEvidence(t *testing.T, feature, role string) {
	t.Helper()
	d := t.TempDir()
	if err := os.MkdirAll(filepath.Join(d, ".workflow"), 0o755); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(d, ".workflow", feature+"-"+role+".json")
	if err := os.WriteFile(p, []byte("{ not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(d)
}

func TestCov2EvidenceAppendSurfacesParseError(t *testing.T) {
	seedCorruptEvidence(t, "feat", "gatekeeper")
	err := runEvidenceAppend(nil, []string{"feat", "gatekeeper", "outputs", "v"})
	if err == nil || !strings.Contains(err.Error(), "evidence parse") {
		t.Fatalf("expected a parse error, got %v", err)
	}
}

func TestCov2EvidenceSetSurfacesParseError(t *testing.T) {
	seedCorruptEvidence(t, "feat", "gatekeeper")
	err := runEvidenceSet(nil, []string{"feat", "gatekeeper", "summary", "v"})
	if err == nil || !strings.Contains(err.Error(), "evidence parse") {
		t.Fatalf("expected a parse error, got %v", err)
	}
}

func TestCov2EvidenceReadSurfacesParseError(t *testing.T) {
	seedCorruptEvidence(t, "feat", "gatekeeper")
	evidenceReadField = ""
	err := runEvidenceRead(nil, []string{"feat", "gatekeeper"})
	if err == nil || !strings.Contains(err.Error(), "evidence parse") {
		t.Fatalf("expected a parse error, got %v", err)
	}
}

// TestCov2EvidenceAppendLockMkdirError forces evidence.Lock's MkdirAll to fail
// by planting a regular file where the .workflow directory must live.
func TestCov2EvidenceAppendLockMkdirError(t *testing.T) {
	d := t.TempDir()
	if err := os.WriteFile(filepath.Join(d, ".workflow"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(d)
	err := runEvidenceAppend(nil, []string{"feat", "gatekeeper", "outputs", "v"})
	if err == nil || !strings.Contains(err.Error(), "lock") {
		t.Fatalf("expected a lock mkdir error, got %v", err)
	}
}

func TestCov2EvidenceSetLockMkdirError(t *testing.T) {
	d := t.TempDir()
	if err := os.WriteFile(filepath.Join(d, ".workflow"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(d)
	err := runEvidenceSet(nil, []string{"feat", "gatekeeper", "summary", "v"})
	if err == nil || !strings.Contains(err.Error(), "lock") {
		t.Fatalf("expected a lock mkdir error, got %v", err)
	}
}

// TestCov2EvidenceRepairSurfacesRemoveError plants a non-empty directory at a
// matched <feature>-<role>.json.tmp path so os.Remove fails (ENOTEMPTY).
func TestCov2EvidenceRepairSurfacesRemoveError(t *testing.T) {
	d := t.TempDir()
	orphan := filepath.Join(d, ".workflow", "feat-gatekeeper.json.tmp")
	if err := os.MkdirAll(filepath.Join(orphan, "child"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(d)
	if err := runEvidenceRepair(nil, []string{"feat"}); err == nil {
		t.Fatal("expected a repair remove error")
	}
}
