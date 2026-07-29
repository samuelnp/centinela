package workflow

import (
	"os"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/orchestration"
)

func TestValidateArtifactsStrictOrchestrationAndLegacy(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)                                        //nolint:errcheck
	os.Chdir(d)                                              //nolint:errcheck
	os.MkdirAll("docs/plans", 0755)                          //nolint:errcheck
	os.WriteFile("docs/plans/f.md", []byte("x"), 0644)       //nolint:errcheck
	os.MkdirAll("docs/features", 0755)                       //nolint:errcheck
	os.WriteFile("docs/features/f.md", []byte("x"), 0644)    //nolint:errcheck
	os.MkdirAll("specs", 0755)                               //nolint:errcheck
	os.WriteFile("specs/f.feature", []byte("Feature"), 0644) //nolint:errcheck
	os.MkdirAll(WorkflowDir, 0755)                           //nolint:errcheck
	wf := New("f")
	Save(wf) //nolint:errcheck
	if err := ValidateArtifacts("f", "plan", &config.Config{}); err == nil {
		t.Fatal("expected strict orchestration evidence failure")
	}
	writePlanEvidence(t, "f", orchestration.RolePlanner, "senior-engineer")
	if err := ValidateArtifacts("f", "plan", &config.Config{}); err != nil {
		t.Fatalf("expected strict plan success, got %v", err)
	}
	wf.OrchestrationMode = ""
	Save(wf)                                                              //nolint:errcheck
	os.Remove(orchestration.MarkdownPath("f", orchestration.RolePlanner)) //nolint:errcheck
	if err := ValidateArtifacts("f", "plan", &config.Config{}); err != nil {
		t.Fatalf("legacy workflow should skip strict gate: %v", err)
	}
}
