package roadmap

import (
	"os"
	"path/filepath"
	"testing"
)

// BootstrapComplete returns false when the roadmap has no Bootstrap phase.
func TestBootstrapComplete_NoBootstrapPhase(t *testing.T) {
	r := &Roadmap{Phases: []Phase{{Name: "Phase 1: Core", Features: []Feature{{Name: "a"}}}}}
	if BootstrapComplete(r) {
		t.Fatal("expected false when there is no Bootstrap phase")
	}
}

// LoadBacklogFinding surfaces a read error for a missing roadmap.json.
func TestLoadBacklogFinding_ReadError(t *testing.T) {
	if _, err := LoadBacklogFinding(filepath.Join(t.TempDir(), "absent.json"), "x"); err == nil {
		t.Fatal("expected read error for missing roadmap.json")
	}
}

// Defer surfaces a read error when roadmap.json is absent (after validation passes).
func TestDefer_ReadError(t *testing.T) {
	err := Defer(filepath.Join(t.TempDir(), "absent.json"), DeferOptions{
		Slug:    "valid-slug",
		Summary: "a valid summary",
	})
	if err == nil {
		t.Fatal("expected read error for missing roadmap.json")
	}
}

// Promote surfaces an unmarshal error when the located finding is malformed.
func TestPromote_FindingUnmarshalError(t *testing.T) {
	p := filepath.Join(t.TempDir(), "roadmap.json")
	// "source" as a string passes featureName (name only) but fails BacklogFinding.
	body := `{"phases":[{"name":"Backlog","features":[{"name":"x","source":"oops"}]}]}`
	if err := os.WriteFile(p, []byte(body), 0644); err != nil {
		t.Fatal(err)
	}
	scores, _ := ParseScores("9,9,9,9,9,9")
	_, err := Promote(p, PromoteRequest{Slug: "x", Phase: "Phase 1", Scores: scores})
	if err == nil {
		t.Fatal("expected unmarshal error for malformed Backlog finding")
	}
}

// appendPromotionArtifacts surfaces the analysis appendFeatureEntry error.
func TestAppendPromotionArtifacts_AnalysisMissing(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.MkdirAll(".workflow", 0755); err != nil {
		t.Fatal(err)
	}
	// RoadmapAnalysisFile is absent -> appendFeatureEntry fails on the first write.
	scores, _ := ParseScores("9,9,9,9,9,9")
	f := &BacklogFinding{Name: "x", Summary: "s", DeferredAt: "t"}
	if err := appendPromotionArtifacts("x", "s", scores, f); err == nil {
		t.Fatal("expected error when analysis artifact is missing")
	}
}
