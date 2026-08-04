package main

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/roadmap"
)

// coldPromoteTree lays the exact tree AC6 blesses — PROJECT.md plus a roadmap
// json with a bootstrap phase, a target phase, a Backlog finding and a draft —
// and NOTHING else. No ROADMAP.md, no analysis, no quality report.
func coldPromoteTree(t *testing.T, toml string) {
	t.Helper()
	t.Chdir(t.TempDir())
	os.WriteFile("PROJECT.md", []byte("Project Stage: greenfield\n"), 0644) //nolint:errcheck
	if toml != "" {
		os.WriteFile("centinela.toml", []byte(toml), 0644) //nolint:errcheck
	}
	roadmap.Save(&roadmap.Roadmap{Phases: []roadmap.Phase{
		{Name: "Phase 0: Bootstrap", Features: []roadmap.Feature{{Name: "setup"}}},
		{Name: "Phase 1", Features: []roadmap.Feature{{Name: "a-draft", Draft: true}}},
		{Name: "Backlog", Features: []roadmap.Feature{{Name: "finding", Summary: "s"}}},
	}}) //nolint:errcheck
}

// TestPromoteRequiresGrading covers the knob both directions, including the
// fail-safe: an unparseable config keeps the artifacts mandatory.
func TestPromoteRequiresGrading(t *testing.T) {
	for _, tc := range []struct {
		name, toml string
		want       bool
	}{
		{"zero config inherits guided", "", false},
		{"explicit strict", "[workflow]\nenforcement_profile = \"strict\"\n", true},
		{"explicit outcome", "[workflow]\nenforcement_profile = \"outcome\"\n", false},
		{"unparseable config fails safe", "not = = toml", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			coldPromoteTree(t, tc.toml)
			t.Setenv("CENTINELA_MODEL", "")
			if got := promoteRequiresGrading(); got != tc.want {
				t.Fatalf("promoteRequiresGrading() = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestGuidedPromoteSeedsAbsentArtifacts is the F3 fix: on a guided cold-start
// tree, the command `start` sends the operator to must actually run.
func TestGuidedPromoteSeedsAbsentArtifacts(t *testing.T) {
	coldPromoteTree(t, "")
	promotePhase, promoteScores, promoteSummary = "Phase 1", "3,3,3,3,3,3", ""
	if err := promoteScored("finding"); err != nil {
		t.Fatalf("guided promote must succeed on a cold tree, got %v", err)
	}
	quality, err := os.ReadFile(roadmap.RoadmapQualityFile)
	if err != nil {
		t.Fatalf("promote must have seeded the quality artifact: %v", err)
	}
	if !strings.Contains(string(quality), `"name":"finding"`) {
		t.Fatalf("the promoted feature must be recorded: %s", quality)
	}
	r, _ := roadmap.Load()
	if roadmap.IsBacklogFeature(r, "finding") {
		t.Fatal("the finding must have left the Backlog")
	}
}

// TestStrictPromoteStillRefusesAbsentArtifacts is the ❌ direction: strict must
// be byte-identical to the pre-fix behavior, refusing before any write.
func TestStrictPromoteStillRefusesAbsentArtifacts(t *testing.T) {
	coldPromoteTree(t, "[workflow]\nenforcement_profile = \"strict\"\n")
	before, _ := os.ReadFile(roadmap.RoadmapFile)
	promotePhase, promoteScores, promoteSummary = "Phase 1", "9,9,9,9,9,9", ""
	err := promoteScored("finding")
	if err == nil || !strings.Contains(err.Error(), "roadmap artifact json missing") {
		t.Fatalf("strict must still refuse a missing artifact, got %v", err)
	}
	if after, _ := os.ReadFile(roadmap.RoadmapFile); string(after) != string(before) {
		t.Fatal("a refused promote must leave roadmap.json byte-identical")
	}
	if _, err := os.Stat(roadmap.RoadmapQualityFile); err == nil {
		t.Fatal("strict must never seed artifacts")
	}
}
