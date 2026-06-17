package main

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/roadmap"
	"github.com/samuelnp/centinela/internal/workflow"
)

// TestWorkflowOrderForFeature_DepBlocked tests line 44-46 in start_guard.go
// (the checkDependencyGuard error path within workflowOrderForFeature).
func TestWorkflowOrderForFeature_DepBlocked(t *testing.T) {
	d := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig)                                                    //nolint:errcheck
	os.Chdir(d)                                                             //nolint:errcheck
	os.WriteFile("PROJECT.md", []byte("Project Stage: greenfield\n"), 0644) //nolint:errcheck
	os.MkdirAll(workflow.WorkflowDir, 0755)                                 //nolint:errcheck

	// Bootstrap complete so we get to checkDependencyGuard
	r := &roadmap.Roadmap{Phases: []roadmap.Phase{
		{Name: "Phase 0: Bootstrap", Features: []roadmap.Feature{{Name: "bootstrap-done"}}},
		{Name: "Phase 1", Features: []roadmap.Feature{
			{Name: "dep-feat"},
			{Name: "blocked-feat", DependsOn: []string{"dep-feat"}},
		}},
	}}
	roadmap.Save(r) //nolint:errcheck
	// Mark bootstrap complete
	seedWF(t, "bootstrap-done", "done")
	writeRoadmapAnalysis(t, "bootstrap-done", "dep-feat", "blocked-feat")
	writeRoadmapQuality(t, 9, "bootstrap-done", "dep-feat", "blocked-feat")

	// blocked-feat has dep-feat (planned) -> blocked
	_, err := workflowOrderForFeature("blocked-feat")
	if err == nil {
		t.Fatal("expected blocked error for dep-blocked feature")
	}
	if !strings.Contains(err.Error(), "dep-feat") {
		t.Errorf("error must name dep, got: %v", err)
	}
}
