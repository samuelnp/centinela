package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Acceptance: specs/brownfield-roadmap-generation.feature

// Scenario: Baseline features are excluded from status counts and validate coverage
func TestAccBrown_BaselineExcludedFromStatusAndCoverage(t *testing.T) {
	dir := brownDir(t, goBrownInventory)
	code := 0
	var stdout string
	if stdout, code = runBrownBin(t, dir); code != 0 {
		t.Fatalf("expected exit 0, got %d:\n%s", code, stdout)
	}
	// Built surfaces land under the Baseline phase-name convention — the same
	// predicate that exempts Backlog excludes Baseline from status/coverage —
	// and the draft never touches the canonical roadmap.json status reads.
	body, _ := os.ReadFile(filepath.Join(dir, ".workflow", "roadmap.brownfield.json"))
	if !strings.Contains(string(body), `"name": "Baseline"`) {
		t.Fatalf("draft must use the Baseline convention so status/coverage exempt it:\n%s", body)
	}
	if _, err := os.Stat(filepath.Join(dir, ".workflow", "roadmap.json")); err == nil {
		t.Fatal("brownfield must not create a canonical roadmap.json that status would count")
	}
}

// Scenario: Running twice on an unchanged inventory yields byte-identical draft output
func TestAccBrown_Deterministic(t *testing.T) {
	dir := brownDir(t, goBrownInventory)
	runBrownBin(t, dir, "--out", "a.json")
	runBrownBin(t, dir, "--out", "b.json")
	a, _ := os.ReadFile(filepath.Join(dir, "a.json"))
	b, _ := os.ReadFile(filepath.Join(dir, "b.json"))
	if len(a) == 0 || string(a) != string(b) {
		t.Fatal("two runs on an unchanged inventory must be byte-identical")
	}
}

// Scenario: Missing inventory fails with guidance and writes nothing
func TestAccBrown_MissingInventoryGuidesAndWritesNothing(t *testing.T) {
	dir := brownDir(t, "")
	stdout, code := runBrownBin(t, dir)
	if code == 0 || !strings.Contains(stdout, "centinela analyze") {
		t.Fatalf("missing inventory must exit non-zero and guide to analyze, got %d:\n%s", code, stdout)
	}
	if _, err := os.Stat(filepath.Join(dir, ".workflow", "roadmap.brownfield.json")); err == nil {
		t.Fatal("no draft must be written on missing-inventory failure")
	}
}

// Scenario: An empty doc-only inventory yields an empty Baseline and zero gaps
func TestAccBrown_DocOnlyEmptyBaselineZeroGaps(t *testing.T) {
	dir := brownDir(t, docOnlyBrownInventory)
	stdout, code := runBrownBin(t, dir)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d:\n%s", code, stdout)
	}
	if !strings.Contains(stdout, "baseline entries: 0") || !strings.Contains(stdout, "gaps: 0") {
		t.Fatalf("doc-only inventory must report 0 baseline 0 gaps:\n%s", stdout)
	}
	body, _ := os.ReadFile(filepath.Join(dir, ".workflow", "roadmap.brownfield.json"))
	if !strings.Contains(string(body), `"name": "Baseline"`) {
		t.Fatalf("draft must not be malformed — a Baseline phase must remain:\n%s", body)
	}
}

// Scenario: A built repo with no TODOs and no goals produces a Baseline-only draft with a hint
func TestAccBrown_BaselineOnlyDraftHasHint(t *testing.T) {
	dir := brownDir(t, docOnlyBrownInventory)
	stdout, _ := runBrownBin(t, dir)
	if !strings.Contains(stdout, "supply --goal") {
		t.Fatalf("a no-gap draft must hint at --goal:\n%s", stdout)
	}
	body, _ := os.ReadFile(filepath.Join(dir, ".workflow", "roadmap.brownfield.json"))
	if strings.Contains(string(body), `"name": "Gaps"`) {
		t.Fatalf("a no-gap draft must carry no Gaps phase:\n%s", body)
	}
}
