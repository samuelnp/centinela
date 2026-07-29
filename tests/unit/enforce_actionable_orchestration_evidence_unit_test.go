package unit_test

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/orchestration"
	"github.com/samuelnp/centinela/internal/workflow"
)

func TestPlanStep_RejectsSummaryOnlyOutputs(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)                                                                                //nolint:errcheck
	os.Chdir(d)                                                                                      //nolint:errcheck
	os.MkdirAll("docs/features", 0755)                                                               //nolint:errcheck
	os.MkdirAll("docs/plans", 0755)                                                                  //nolint:errcheck
	os.MkdirAll("specs", 0755)                                                                       //nolint:errcheck
	os.MkdirAll(workflow.WorkflowDir, 0755)                                                          //nolint:errcheck
	os.WriteFile("docs/features/f.md", []byte("brief"), 0644)                                        //nolint:errcheck
	os.WriteFile("docs/plans/f.md", []byte("plan"), 0644)                                            //nolint:errcheck
	os.WriteFile("specs/f.feature", []byte("Feature: f"), 0644)                                      //nolint:errcheck
	workflow.Save(workflow.New("f"))                                                                 //nolint:errcheck
	os.WriteFile(orchestration.MarkdownPath("f", orchestration.RolePlanner), []byte("# role"), 0644) //nolint:errcheck
	ts := time.Now().UTC().Format(time.RFC3339)
	// A summary string instead of a real plan/spec path must still be rejected.
	plan := `{"feature":"f","step":"plan","role":"planner","status":"done","generatedAt":"` + ts + `","inputs":["docs/features/f.md","docs/plans/f.md"],"outputs":["plan approved"],"edgeCases":["e"],"handoffTo":"senior-engineer"}`
	os.WriteFile(orchestration.JSONPath("f", orchestration.RolePlanner), []byte(plan), 0644) //nolint:errcheck
	err := workflow.ValidateArtifacts("f", "plan", &config.Config{})
	if err == nil || !strings.Contains(err.Error(), "actionable outputs") {
		t.Fatalf("expected actionable output error, got %v", err)
	}
}
