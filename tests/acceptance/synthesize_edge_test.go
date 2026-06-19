package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Acceptance: specs/archetype-inference-project-synthesis.feature

// Scenario: An ambiguous inventory is flagged low-confidence with a rationale
func TestAccSynth_AmbiguousLowConfidence(t *testing.T) {
	dir := t.TempDir()
	in := writeAnalysis(t, dir, ambiguousInventory)
	out, code := runSynthesizeBin(t, dir, "--in", in, "--out", filepath.Join(dir, "PROJECT.md"))
	if code != 0 || !strings.Contains(out, "confidence: low") || !strings.Contains(out, "ambiguous") {
		t.Fatalf("expected ambiguous low confidence (code %d):\n%s", code, out)
	}
	body, _ := os.ReadFile(filepath.Join(dir, "PROJECT.md"))
	if !strings.Contains(string(body), "DRAFT") {
		t.Fatalf("draft must be marked DRAFT:\n%s", body)
	}
}

// Scenario: Running synthesize without an inventory fails with guidance
func TestAccSynth_MissingInventoryFails(t *testing.T) {
	dir := t.TempDir()
	out, code := runSynthesizeBin(t, dir, "--in", filepath.Join(dir, "absent.json"), "--out", filepath.Join(dir, "PROJECT.md"))
	if code == 0 || !strings.Contains(out, "centinela analyze") {
		t.Fatalf("missing inventory must fail with guidance (code %d):\n%s", code, out)
	}
	if _, err := os.Stat(filepath.Join(dir, "PROJECT.md")); err == nil {
		t.Fatal("no PROJECT.md should be written on failure")
	}
}

// Scenario: An existing PROJECT.md is preserved and a draft is written instead
func TestAccSynth_ExistingPreserved(t *testing.T) {
	dir := t.TempDir()
	in := writeAnalysis(t, dir, railsInventory)
	writeFile(t, dir, "PROJECT.md", "ORIGINAL")
	out, code := runSynthesizeBin(t, dir, "--in", in, "--out", filepath.Join(dir, "PROJECT.md"))
	if code != 0 || !strings.Contains(out, "preserved") {
		t.Fatalf("existing PROJECT.md must be preserved (code %d):\n%s", code, out)
	}
	if b, _ := os.ReadFile(filepath.Join(dir, "PROJECT.md")); string(b) != "ORIGINAL" {
		t.Fatalf("original mutated: %q", b)
	}
	if _, err := os.Stat(filepath.Join(dir, "PROJECT.draft.md")); err != nil {
		t.Fatalf("PROJECT.draft.md not written: %v", err)
	}
}
