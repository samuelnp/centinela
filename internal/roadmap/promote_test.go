package roadmap

import (
	"os"
	"strings"
	"testing"
)

// setupPromoteDir creates a temp dir with roadmap.json and all artifacts.
func setupPromoteDir(t *testing.T) string {
	t.Helper()
	d := t.TempDir()
	orig, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(orig) }) //nolint:errcheck
	os.Chdir(d)                          //nolint:errcheck
	os.MkdirAll(".workflow", 0755)       //nolint:errcheck
	roadmap := `{"phases":[{"name":"Phase 5","features":[]},{"name":"Backlog","features":[{"name":"hook-timeout-config","summary":"Timeout hardcoded","source":{"feature":"dfrc","role":"senior-engineer"},"deferredAt":"2026-01-01T00:00:00Z"}]}]}`
	os.WriteFile(RoadmapFile, []byte(roadmap), 0644)                                                                   //nolint:errcheck
	os.WriteFile(RoadmapAnalysisFile, []byte(`{"role":"senior-product-manager","features":[]}`), 0644)                 //nolint:errcheck
	os.WriteFile(RoadmapQualityFile, []byte(`{"role":"roadmap-quality-evaluator","threshold":9,"features":[]}`), 0644) //nolint:errcheck
	os.WriteFile(RoadmapAnalysisMarkdown, []byte("# analysis\n"), 0644)                                                //nolint:errcheck
	os.WriteFile(RoadmapQualityMarkdown, []byte("# quality\n"), 0644)                                                  //nolint:errcheck
	return d
}

// TestLoadBacklogFinding_Found decodes the finding from Backlog.
func TestLoadBacklogFinding_Found(t *testing.T) {
	setupPromoteDir(t)
	f, err := LoadBacklogFinding(RoadmapFile, "hook-timeout-config")
	if err != nil {
		t.Fatalf("LoadBacklogFinding: %v", err)
	}
	if f.Name != "hook-timeout-config" || f.Summary == "" {
		t.Errorf("unexpected finding: %+v", f)
	}
	if f.Source == nil || f.Source.Feature != "dfrc" {
		t.Errorf("source not decoded: %+v", f)
	}
}

// TestLoadBacklogFinding_NotFound returns error.
func TestLoadBacklogFinding_NotFound(t *testing.T) {
	setupPromoteDir(t)
	if _, err := LoadBacklogFinding(RoadmapFile, "nonexistent"); err == nil {
		t.Error("expected error for missing Backlog slug")
	}
}

// TestPromote_HappyPath moves slug and appends artifacts.
func TestPromote_HappyPath(t *testing.T) {
	setupPromoteDir(t)
	scores, _ := ParseScores("9,9,8,7,9,9")
	f, err := Promote(RoadmapFile, PromoteRequest{
		Slug: "hook-timeout-config", Phase: "Phase 5", Scores: scores,
	})
	if err != nil {
		t.Fatalf("Promote: %v", err)
	}
	if f == nil || f.Name != "hook-timeout-config" {
		t.Errorf("returned finding unexpected: %v", f)
	}
	roadmapData, _ := os.ReadFile(RoadmapFile)
	s := string(roadmapData)
	if !strings.Contains(s, "Phase 5") {
		t.Error("Phase 5 must be present")
	}
	if !strings.Contains(s, "hook-timeout-config") {
		t.Error("slug must appear in output roadmap")
	}
	analysisData, _ := os.ReadFile(RoadmapAnalysisFile)
	if !strings.Contains(string(analysisData), "hook-timeout-config") {
		t.Error("analysis must contain promoted slug")
	}
}
