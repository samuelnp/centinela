package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Acceptance: specs/spec-reconstruction.feature

// Scenario: A valid inventory reconstructs feature skeletons and brief stubs into the review dir
func TestAccRecon_ValidInventoryWritesCorpus(t *testing.T) {
	dir := t.TempDir()
	in := writeAnalysis(t, dir, goNtierReconInventory)
	out := filepath.Join(dir, "review")
	stdout, code := runReconstructBin(t, dir, "--in", in, "--out", out)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d:\n%s", code, stdout)
	}
	specs, _ := filepath.Glob(filepath.Join(out, "specs", "*.feature"))
	briefs, _ := filepath.Glob(filepath.Join(out, "features", "*.md"))
	if len(specs) == 0 || len(briefs) == 0 {
		t.Fatalf("expected feature + brief stubs, got %d specs %d briefs", len(specs), len(briefs))
	}
}

// Scenario: Every generated feature parses with the spec traceability scenario parser
func TestAccRecon_GeneratedFeaturesParse(t *testing.T) {
	dir := t.TempDir()
	in := writeAnalysis(t, dir, goNtierReconInventory)
	out := filepath.Join(dir, "review")
	runReconstructBin(t, dir, "--in", in, "--out", out)
	specs, _ := filepath.Glob(filepath.Join(out, "specs", "*.feature"))
	if len(specs) == 0 {
		t.Fatal("no feature files generated")
	}
	for _, s := range specs {
		body, _ := os.ReadFile(s)
		if !strings.Contains(string(body), "Feature:") || !strings.Contains(string(body), "Scenario:") {
			t.Fatalf("%s is not valid Gherkin:\n%s", s, body)
		}
	}
}

// Scenario: Unknowable behavior is emitted as an explicit TODO confirm and never fabricated
func TestAccRecon_TodoMarkersPresent(t *testing.T) {
	dir := t.TempDir()
	in := writeAnalysis(t, dir, goNtierReconInventory)
	out := filepath.Join(dir, "review")
	runReconstructBin(t, dir, "--in", in, "--out", out)
	specs, _ := filepath.Glob(filepath.Join(out, "specs", "*.feature"))
	body, _ := os.ReadFile(specs[0])
	if !strings.Contains(string(body), "# TODO: confirm") {
		t.Fatalf("expected explicit TODO confirm markers:\n%s", body)
	}
}

// Scenario: Re-running reconstruct on an unchanged inventory produces byte-identical output
func TestAccRecon_Deterministic(t *testing.T) {
	dir := t.TempDir()
	in := writeAnalysis(t, dir, goNtierReconInventory)
	runReconstructBin(t, dir, "--in", in, "--out", filepath.Join(dir, "a"))
	runReconstructBin(t, dir, "--in", in, "--out", filepath.Join(dir, "b"))
	specs, _ := filepath.Glob(filepath.Join(dir, "a", "specs", "*.feature"))
	for _, fa := range specs {
		fb := strings.Replace(fa, filepath.Join(dir, "a"), filepath.Join(dir, "b"), 1)
		ba, _ := os.ReadFile(fa)
		bb, _ := os.ReadFile(fb)
		if string(ba) != string(bb) || len(ba) == 0 {
			t.Fatalf("reconstruct output not byte-identical for %s", filepath.Base(fa))
		}
	}
}

// Scenario: The summary reports targets selected files written and total TODO markers
func TestAccRecon_SummaryReports(t *testing.T) {
	dir := t.TempDir()
	in := writeAnalysis(t, dir, goNtierReconInventory)
	stdout, _ := runReconstructBin(t, dir, "--in", in, "--out", filepath.Join(dir, "review"))
	for _, want := range []string{"targets selected:", "files written:", "TODO confirm markers:"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("summary missing %q:\n%s", want, stdout)
		}
	}
}
