package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/roadmap"
	"github.com/samuelnp/centinela/internal/workflow"
)

// TestCov2DoneCountUnknownStep drives the fall-through return 0 when the
// current step is not one of the ordered steps.
func TestCov2DoneCountUnknownStep(t *testing.T) {
	wf := workflow.New("alpha")
	wf.CurrentStep = "not-a-real-step"
	if got := doneCount(wf); got != 0 {
		t.Fatalf("unknown current step must yield 0, got %d", got)
	}
}

// TestCov2RoadmapIterateWriteMarkerError forces WriteMarker's MkdirAll to fail
// by planting a regular file where the .workflow directory must be created.
func TestCov2RoadmapIterateWriteMarkerError(t *testing.T) {
	d := t.TempDir()
	if err := os.WriteFile(filepath.Join(d, ".workflow"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(d)
	if err := runRoadmapIterate(nil, nil); err == nil {
		t.Fatal("expected WriteMarker mkdir error")
	}
}

// TestCov2RoadmapValidateQualityMissing covers the ValidateQuality error branch:
// analysis artifacts present and valid, but the quality artifacts are absent.
func TestCov2RoadmapValidateQualityMissing(t *testing.T) {
	chdirIntoTemp(t)
	r := &roadmap.Roadmap{Phases: []roadmap.Phase{{Name: "P1", Features: []roadmap.Feature{{Name: "user"}}}}}
	if err := roadmap.Save(r); err != nil {
		t.Fatal(err)
	}
	writeRoadmapAnalysis(t, "user")
	err := runRoadmapValidate(nil, nil)
	if err == nil || !strings.Contains(err.Error(), "quality") {
		t.Fatalf("expected a quality-missing error, got %v", err)
	}
}

// TestCov2ShouldRenderReviewPromptNilWorkflow covers the nil/done short-circuit.
func TestCov2ShouldRenderReviewPromptNilWorkflow(t *testing.T) {
	if shouldRenderReviewPrompt(nil, &config.Config{}) {
		t.Fatal("a nil workflow must never render the review prompt")
	}
}

// TestCov2SyncWorktreeIgnoresError surfaces an underlying SyncIgnores failure
// (tsconfig.json present as a directory makes the read fail).
func TestCov2SyncWorktreeIgnoresError(t *testing.T) {
	d := t.TempDir()
	if err := os.Mkdir(filepath.Join(d, "tsconfig.json"), 0o755); err != nil {
		t.Fatal(err)
	}
	err := syncWorktreeIgnores(d)
	if err == nil || !strings.Contains(err.Error(), "worktree ignore sync failed") {
		t.Fatalf("expected a wrapped sync error, got %v", err)
	}
}
