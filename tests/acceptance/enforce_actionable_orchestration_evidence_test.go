package acceptance_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/samuelnp/centinela/internal/orchestration"
	"github.com/samuelnp/centinela/internal/workflow"
)

// Scenario: Code evidence fails without real implementation outputs
func TestCompleteFailsForInsightOnlyCodeEvidence(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	repo := filepath.Clean(filepath.Join(o, "..", ".."))
	bin := filepath.Join(d, "centinela-test")
	build := exec.Command("go", "build", "-o", bin, "./cmd/centinela")
	build.Dir = repo
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build centinela failed: %v\n%s", err, out)
	}
	os.Chdir(d)                             //nolint:errcheck
	os.MkdirAll(workflow.WorkflowDir, 0755) //nolint:errcheck
	wf := workflow.New("f")
	wf.CurrentStep = "code"
	workflow.Save(wf)                                                                                       //nolint:errcheck
	os.WriteFile(orchestration.MarkdownPath("f", orchestration.RoleSeniorEngineer), []byte("# role"), 0644) //nolint:errcheck
	os.WriteFile(".workflow/f-senior-engineer.md", []byte("evidence"), 0644)                                //nolint:errcheck
	ts := time.Now().UTC().Format(time.RFC3339)
	data := `{"feature":"f","step":"code","role":"senior-engineer","status":"done","generatedAt":"` + ts + `","inputs":["docs/plans/f.md"],"outputs":[".workflow/f-senior-engineer.md"],"edgeCases":[],"handoffTo":"qa-senior"}`
	os.WriteFile(orchestration.JSONPath("f", orchestration.RoleSeniorEngineer), []byte(data), 0644) //nolint:errcheck
	cmd := exec.Command(bin, "complete", "f")
	cmd.Dir = d
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected completion failure, got success: %s", out)
	}
	if !strings.Contains(string(out), "non-evidence implementation file") {
		t.Fatalf("expected actionable failure details, got: %s", out)
	}
}
