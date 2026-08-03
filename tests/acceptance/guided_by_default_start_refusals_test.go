// Acceptance: specs/guided-by-default.feature
package acceptance_test

import (
	"path/filepath"
	"strings"
	"testing"
)

// Scenario: Guided still refuses what the roadmap says it must
//
// Covers three of the outline's five examples (Backlog finding, draft
// feature, no bootstrap phase); unmet dependencies and an incomplete
// bootstrap are exercised at unit level in
// cmd/centinela/start_guard_guided_refusals_test.go.
func TestGBD_GuidedStillRefusesWhatRoadmapRequires(t *testing.T) {
	bin := avvBuildBin(t)
	cases := []struct {
		name     string
		roadmap  string
		feature  string
		wantText string
	}{
		{
			name: "Backlog finding",
			roadmap: `{"phases":[{"name":"Phase 0: Bootstrap","features":[]},` +
				`{"name":"Backlog","features":[{"name":"a-finding"}]}]}`,
			feature: "a-finding", wantText: "Backlog",
		},
		{
			name: "draft feature",
			roadmap: `{"phases":[{"name":"Phase 0: Bootstrap",` +
				`"features":[{"name":"a-draft","draft":true}]}]}`,
			feature: "a-draft", wantText: "draft",
		},
		{
			name:    "no bootstrap phase",
			roadmap: `{"phases":[{"name":"Phase 1: Foundations","features":[{"name":"x"}]}]}`,
			feature: "x", wantText: "Phase 0: Bootstrap",
		},
	}
	for _, tc := range cases {
		dir := t.TempDir()
		mustWrite(t, filepath.Join(dir, "PROJECT.md"), "Project Stage: greenfield\n")
		mustWrite(t, filepath.Join(dir, ".workflow", "roadmap.json"), tc.roadmap)
		out, code := runCent(t, bin, dir, "start", tc.feature)
		if code == 0 {
			t.Fatalf("%s: guided must still refuse, got exit 0: %s", tc.name, out)
		}
		if !strings.Contains(out, tc.wantText) {
			t.Fatalf("%s: refusal must mention %q, got: %s", tc.name, tc.wantText, out)
		}
	}
}

// Scenario: A missing roadmap json is still refused under guided
func TestGBD_MissingRoadmapJSONRefusedUnderGuided(t *testing.T) {
	bin := avvBuildBin(t)
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "PROJECT.md"), "Project Stage: greenfield\n")
	out, code := runCent(t, bin, dir, "start", "anything")
	if code == 0 {
		t.Fatalf("a missing roadmap json must refuse, got exit 0: %s", out)
	}
	if !strings.Contains(out, "roadmap.json") {
		t.Fatalf("refusal must name the roadmap json: %s", out)
	}
}
