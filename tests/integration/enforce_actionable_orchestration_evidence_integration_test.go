package integration_test

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/orchestration"
	"github.com/samuelnp/centinela/internal/workflow"
)

func TestCodeStep_RequiresRealImplementationOutput(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)                       //nolint:errcheck
	os.Chdir(d)                             //nolint:errcheck
	os.MkdirAll(workflow.WorkflowDir, 0755) //nolint:errcheck
	os.MkdirAll("internal/demo", 0755)      //nolint:errcheck
	wf := workflow.New("f")
	wf.CurrentStep = "code"
	workflow.Save(wf)                                                                                       //nolint:errcheck
	os.WriteFile(orchestration.MarkdownPath("f", orchestration.RoleSeniorEngineer), []byte("# role"), 0644) //nolint:errcheck
	os.WriteFile(".workflow/f-senior-engineer.md", []byte("evidence"), 0644)                                //nolint:errcheck
	ts := time.Now().UTC().Format(time.RFC3339)
	bad := `{"feature":"f","step":"code","role":"senior-engineer","status":"done","generatedAt":"` + ts + `","inputs":["docs/plans/f.md"],"outputs":[".workflow/f-senior-engineer.md"],"edgeCases":[],"handoffTo":"qa-senior"}`
	os.WriteFile(orchestration.JSONPath("f", orchestration.RoleSeniorEngineer), []byte(bad), 0644) //nolint:errcheck
	err := workflow.ValidateArtifacts("f", "code", &config.Config{})
	if err == nil || !strings.Contains(err.Error(), "non-evidence implementation") {
		t.Fatalf("expected implementation output error, got %v", err)
	}
	os.WriteFile("internal/demo/file.go", []byte("package demo"), 0644) //nolint:errcheck
	good := `{"feature":"f","step":"code","role":"senior-engineer","status":"done","generatedAt":"` + ts + `","inputs":["docs/plans/f.md"],"outputs":["internal/demo/file.go"],"edgeCases":[],"handoffTo":"qa-senior"}`
	os.WriteFile(orchestration.JSONPath("f", orchestration.RoleSeniorEngineer), []byte(good), 0644) //nolint:errcheck
	if err := workflow.ValidateArtifacts("f", "code", &config.Config{}); err != nil {
		t.Fatalf("expected code step success, got %v", err)
	}
}
