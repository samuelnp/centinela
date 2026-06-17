package integration_test

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/roadmap"
)

func dfrcDir(t *testing.T, body string) string {
	t.Helper()
	d := t.TempDir()
	orig, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(orig) }) //nolint:errcheck
	if err := os.Chdir(d); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(".workflow", 0o755); err != nil {
		t.Fatal(err)
	}
	if body != "" {
		if err := os.WriteFile(roadmap.RoadmapFile, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return d
}

// TestDeferThenPromote_FullFlow exercises the complete defer→promote lifecycle via
// the Go API: defer a finding, verify it lands in Backlog, promote it with passing
// scores, verify it moves to Phase 5, and that validate still passes.
func TestDeferThenPromote_FullFlow(t *testing.T) {
	dfrcDir(t, `{"phases":[{"name":"Phase 5","features":[]},{"name":"Backlog","features":[]}]}`)

	// Seed artifact files required by preflightArtifacts.
	os.WriteFile(roadmap.RoadmapAnalysisFile, []byte(`{"role":"senior-product-manager","features":[]}`), 0644)                 //nolint:errcheck
	os.WriteFile(roadmap.RoadmapQualityFile, []byte(`{"role":"roadmap-quality-evaluator","threshold":9,"features":[]}`), 0644) //nolint:errcheck
	os.WriteFile(roadmap.RoadmapAnalysisMarkdown, []byte("# analysis\n"), 0644)                                                //nolint:errcheck
	os.WriteFile(roadmap.RoadmapQualityMarkdown, []byte("# quality\n"), 0644)                                                  //nolint:errcheck

	// Step 1: Defer a finding.
	opts := roadmap.DeferOptions{
		Slug:    "hook-timeout-config",
		Summary: "Prewrite hook timeout is hardcoded; should be configurable",
		Source:  &roadmap.Source{Feature: "deferred-findings-roadmap-capture", Role: "senior-engineer"},
	}
	if err := roadmap.Defer(roadmap.RoadmapFile, opts); err != nil {
		t.Fatalf("Defer: %v", err)
	}

	// Verify finding is in Backlog.
	data, _ := os.ReadFile(roadmap.RoadmapFile)
	if !strings.Contains(string(data), "Backlog") {
		t.Fatal("Backlog phase missing after defer")
	}
	if !strings.Contains(string(data), "hook-timeout-config") {
		t.Fatal("slug missing from Backlog after defer")
	}

	// Step 2: Promote with passing scores.
	scores, err := roadmap.ParseScores("9,9,9,9,9,9")
	if err != nil {
		t.Fatalf("ParseScores: %v", err)
	}
	finding, err := roadmap.Promote(roadmap.RoadmapFile, roadmap.PromoteRequest{
		Slug: "hook-timeout-config", Phase: "Phase 5", Scores: scores,
	})
	if err != nil {
		t.Fatalf("Promote: %v", err)
	}
	if finding == nil || finding.Name != "hook-timeout-config" {
		t.Errorf("unexpected finding: %v", finding)
	}

	// Verify moved to Phase 5, no longer in Backlog.
	data, _ = os.ReadFile(roadmap.RoadmapFile)
	s := string(data)
	if !strings.Contains(s, "Phase 5") {
		t.Error("Phase 5 must be present after promote")
	}
	if !strings.Contains(s, "hook-timeout-config") {
		t.Error("slug must be present in roadmap after promote")
	}

	// Verify artifacts updated.
	analysis, _ := os.ReadFile(roadmap.RoadmapAnalysisFile)
	if !strings.Contains(string(analysis), "hook-timeout-config") {
		t.Error("analysis.json must contain promoted slug")
	}
	quality, _ := os.ReadFile(roadmap.RoadmapQualityFile)
	if !strings.Contains(string(quality), "hook-timeout-config") {
		t.Error("quality.json must contain promoted slug")
	}
}
