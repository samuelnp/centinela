package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Acceptance: specs/brownfield-roadmap-generation.feature

// Scenario: A built repo produces a draft with a Baseline phase listing already-built surfaces
func TestAccBrown_BaselinePhaseListsBuiltSurfaces(t *testing.T) {
	dir := brownDir(t, goBrownInventory)
	stdout, code := runBrownBin(t, dir)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d:\n%s", code, stdout)
	}
	body, _ := os.ReadFile(filepath.Join(dir, ".workflow", "roadmap.brownfield.json"))
	if !strings.Contains(string(body), `"name": "Baseline"`) || !strings.Contains(string(body), "-confirm") {
		t.Fatalf("draft missing a populated Baseline phase:\n%s", body)
	}
}

// Scenario: The command never clobbers an existing canonical roadmap.json
func TestAccBrown_NeverClobbersCanonical(t *testing.T) {
	dir := brownDir(t, goBrownInventory)
	const curated = `{"phases":[{"name":"Phase 1","features":[{"name":"hand-authored"}]}]}`
	canonical := filepath.Join(dir, ".workflow", "roadmap.json")
	if err := os.WriteFile(canonical, []byte(curated), 0o644); err != nil {
		t.Fatal(err)
	}
	stdout, code := runBrownBin(t, dir)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d:\n%s", code, stdout)
	}
	after, _ := os.ReadFile(canonical)
	if string(after) != curated {
		t.Fatal("canonical roadmap.json must be byte-for-byte unchanged")
	}
	if !strings.Contains(stdout, "roadmap.brownfield.json") {
		t.Fatalf("summary must report the draft path:\n%s", stdout)
	}
}

// Scenario: Gap phase — reconstruct TODO confirm markers become net-new gap features
func TestAccBrown_TodoTargetsBecomeGapFeatures(t *testing.T) {
	dir := brownDir(t, goBrownInventory)
	runBrownBin(t, dir)
	body, _ := os.ReadFile(filepath.Join(dir, ".workflow", "roadmap.brownfield.json"))
	if !strings.Contains(string(body), `"name": "Gaps"`) {
		t.Fatalf("expected a Gaps phase distinct from Baseline:\n%s", body)
	}
	if strings.Count(string(body), "-confirm") < 2 {
		t.Fatalf("each TODO-bearing target must become a -confirm gap feature:\n%s", body)
	}
}

// Scenario: A user-stated goal adds a net-new gap feature
func TestAccBrown_GoalAddsGapFeature(t *testing.T) {
	dir := brownDir(t, goBrownInventory)
	runBrownBin(t, dir, "--goal", "Add OAuth login")
	body, _ := os.ReadFile(filepath.Join(dir, ".workflow", "roadmap.brownfield.json"))
	s := string(body)
	gapsIdx := strings.Index(s, `"name": "Gaps"`)
	goalIdx := strings.Index(s, "Add OAuth login")
	if gapsIdx < 0 || goalIdx < gapsIdx {
		t.Fatalf("goal-derived feature must live in the gap phase, not Baseline:\n%s", s)
	}
}

// Scenario: The summary reports baseline count gap count and draft path
func TestAccBrown_SummaryReportsCountsAndPath(t *testing.T) {
	dir := brownDir(t, goBrownInventory)
	stdout, _ := runBrownBin(t, dir)
	for _, want := range []string{"baseline entries:", "gaps:", "draft written:"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("summary missing %q:\n%s", want, stdout)
		}
	}
}
