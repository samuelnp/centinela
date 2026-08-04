package main

import (
	"os"
	"testing"

	"github.com/samuelnp/centinela/internal/roadmap"
)

// TestFailedGuidedPromoteWritesNothing: seeding happens INSIDE Promote, after
// the slug and phase are validated, so a bad request leaves the tree exactly as
// it found it — the "rejected before any write" posture promote otherwise keeps.
func TestFailedGuidedPromoteWritesNothing(t *testing.T) {
	for _, tc := range []struct{ name, slug, phase string }{
		{"unknown slug", "no-such-slug", "Phase 1"},
		{"unknown phase", "finding", "Phase 99"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			coldPromoteTree(t, "")
			before, _ := os.ReadFile(roadmap.RoadmapFile)
			promotePhase, promoteScores, promoteSummary = tc.phase, "9,9,9,9,9,9", ""
			if err := promoteScored(tc.slug); err == nil {
				t.Fatal("an invalid promote must be refused")
			}
			for _, p := range []string{
				roadmap.RoadmapAnalysisFile, roadmap.RoadmapQualityFile,
				roadmap.RoadmapAnalysisMarkdown, roadmap.RoadmapQualityMarkdown,
			} {
				if _, err := os.Stat(p); err == nil {
					t.Errorf("a refused promote must not seed %s", p)
				}
			}
			if after, _ := os.ReadFile(roadmap.RoadmapFile); string(after) != string(before) {
				t.Error("a refused promote must leave roadmap.json byte-identical")
			}
		})
	}
}
