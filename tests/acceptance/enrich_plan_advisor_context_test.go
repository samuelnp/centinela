package acceptance_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

// Acceptance: specs/enrich-plan-advisor-context.feature
func TestPlanAdvisorHookUsesDependencyFirstContext(t *testing.T) {
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
	os.Chdir(d)                                                                                                                                                                                                  //nolint:errcheck
	os.MkdirAll(workflow.WorkflowDir, 0755)                                                                                                                                                                      //nolint:errcheck
	os.MkdirAll("docs/features", 0755)                                                                                                                                                                           //nolint:errcheck
	os.WriteFile("docs/features/f.md", []byte("## Problem\ntext\n"), 0644)                                                                                                                                       //nolint:errcheck
	os.WriteFile("docs/features/dep.md", []byte("RAW DEP BRIEF SHOULD NOT APPEAR"), 0644)                                                                                                                        //nolint:errcheck
	os.WriteFile(".workflow/roadmap.json", []byte(`{"phases":[{"name":"P1","features":[{"name":"dep"},{"name":"sib"},{"name":"f"}]}]}`), 0644)                                                                   //nolint:errcheck
	os.WriteFile(".workflow/roadmap-analysis.json", []byte(`{"role":"senior-product-manager","features":[{"name":"dep","dependsOn":[]},{"name":"sib","dependsOn":[]},{"name":"f","dependsOn":["dep"]}]}`), 0644) //nolint:errcheck
	os.WriteFile(".workflow/dep-edge-cases.md", []byte("- duplicate webhook retries"), 0644)                                                                                                                     //nolint:errcheck
	workflow.Save(workflow.New("f"))                                                                                                                                                                             //nolint:errcheck
	cmd := exec.Command(bin, "hook", "plan-advisor")
	cmd.Dir = d
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("plan-advisor hook failed: %v\n%s", err, out)
	}
	s := string(out)
	if strings.Index(s, "dependencies first: dep") > strings.Index(s, "same-phase siblings: sib") {
		t.Fatalf("expected dependency context before siblings, got: %s", s)
	}
	if !strings.Contains(s, "related edge-case lessons: dep: duplicate webhook retries") {
		t.Fatalf("expected related edge-case lesson, got: %s", s)
	}
	if strings.Contains(s, "RAW DEP BRIEF SHOULD NOT APPEAR") {
		t.Fatalf("expected summarized context only, got: %s", s)
	}
}
