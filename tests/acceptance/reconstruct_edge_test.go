package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Acceptance: specs/spec-reconstruction.feature

// Scenario: A hand-authored spec is never clobbered and is reported as skipped
func TestAccRecon_SkipsHandAuthoredSpec(t *testing.T) {
	dir := t.TempDir()
	in := writeAnalysis(t, dir, goNtierReconInventory)
	// internal/service slugifies to internal-service; pre-create its canonical spec.
	writeFile(t, dir, filepath.Join("specs", "internal-service.feature"),
		"Feature: hand authored\n  Scenario: keep\n")
	orig, _ := os.ReadFile(filepath.Join(dir, "specs", "internal-service.feature"))
	stdout, code := runReconstructBin(t, dir, "--in", in, "--out", filepath.Join(dir, "review"))
	if code != 0 || !strings.Contains(stdout, "skipped") {
		t.Fatalf("expected skip report (code %d):\n%s", code, stdout)
	}
	after, _ := os.ReadFile(filepath.Join(dir, "specs", "internal-service.feature"))
	if string(orig) != string(after) {
		t.Fatal("hand-authored spec must be left byte-for-byte unchanged")
	}
}

// Scenario: Running reconstruct without an inventory fails with guidance and writes nothing
func TestAccRecon_NoInventoryFails(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "review")
	stdout, code := runReconstructBin(t, dir, "--in", filepath.Join(dir, "missing.json"), "--out", out)
	if code == 0 {
		t.Fatalf("expected non-zero exit without an inventory:\n%s", stdout)
	}
	if !strings.Contains(stdout, "centinela analyze") {
		t.Fatalf("expected guidance to run `centinela analyze`:\n%s", stdout)
	}
	if _, err := os.Stat(out); !os.IsNotExist(err) {
		t.Fatal("no corpus must be written when the inventory is missing")
	}
}

// Scenario: An empty doc-only inventory selects zero targets and writes no empty feature
func TestAccRecon_DocOnlyZeroTargets(t *testing.T) {
	dir := t.TempDir()
	in := writeAnalysis(t, dir, docOnlyReconInventory)
	out := filepath.Join(dir, "review")
	stdout, code := runReconstructBin(t, dir, "--in", in, "--out", out)
	if code != 0 || !strings.Contains(stdout, "targets selected: 0") {
		t.Fatalf("expected 0 targets (code %d):\n%s", code, stdout)
	}
	specs, _ := filepath.Glob(filepath.Join(out, "specs", "*.feature"))
	if len(specs) != 0 {
		t.Fatalf("doc-only inventory must write no feature files, got %d", len(specs))
	}
}

// Scenario: A polyglot inventory with an empty Go graph still selects manifest and package targets
func TestAccRecon_PolyglotEmptyGraph(t *testing.T) {
	dir := t.TempDir()
	in := writeAnalysis(t, dir, polyglotReconInventory)
	out := filepath.Join(dir, "review")
	stdout, code := runReconstructBin(t, dir, "--in", in, "--out", out)
	if code != 0 {
		t.Fatalf("expected exit 0 (code %d):\n%s", code, stdout)
	}
	specs, _ := filepath.Glob(filepath.Join(out, "specs", "*.feature"))
	if len(specs) == 0 {
		t.Fatalf("polyglot inventory must still select targets:\n%s", stdout)
	}
}
