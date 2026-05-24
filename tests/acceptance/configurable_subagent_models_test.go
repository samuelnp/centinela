package acceptance_test

// Acceptance: specs/configurable-subagent-models.feature

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

func buildModelsTestBinary(t *testing.T, destDir string) string {
	t.Helper()
	// Capture repo root from current working dir before any chdir.
	orig, _ := os.Getwd()
	repo := filepath.Clean(filepath.Join(orig, "..", ".."))
	bin := filepath.Join(destDir, "centinela-models-test")
	build := exec.Command("go", "build", "-o", bin, "./cmd/centinela")
	build.Dir = repo
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}
	return bin
}

func setupModelsRepo(t *testing.T, tomlBody string) (dir, bin string) {
	t.Helper()
	bin = buildModelsTestBinary(t, t.TempDir())
	d := t.TempDir()
	orig, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(orig) })    //nolint:errcheck
	os.Chdir(d)                             //nolint:errcheck
	os.MkdirAll(workflow.WorkflowDir, 0755) //nolint:errcheck
	wf := workflow.New("myfeature")
	wf.CurrentStep = "plan"
	workflow.Save(wf) //nolint:errcheck
	if tomlBody != "" {
		os.WriteFile("centinela.toml", []byte(tomlBody), 0644) //nolint:errcheck
	}
	return d, bin
}

// AC1 + AC6: configured tier annotates the role; both-runner reference emitted.
func TestOrchestrationHook_ConfiguredTierAnnotated(t *testing.T) {
	d, bin := setupModelsRepo(t, "[orchestration.models]\nbig-thinker = \"reasoning\"\n")
	cmd := exec.Command(bin, "hook", "orchestration")
	cmd.Dir = d
	out, _ := cmd.CombinedOutput()
	s := string(out)
	if !strings.Contains(s, "big-thinker (model: reasoning)") {
		t.Errorf("AC1: expected annotation; got:\n%s", s)
	}
	if !strings.Contains(s, "claude-opus-4-7") || !strings.Contains(s, "anthropic/claude-opus-4-7") {
		t.Errorf("AC6: expected both-runner IDs; got:\n%s", s)
	}
	if !strings.Contains(s, "model reference:") {
		t.Errorf("AC6: expected model reference line; got:\n%s", s)
	}
}

// AC2 + AC3: absent table — every role uses its default tier.
func TestOrchestrationHook_AbsentTableAllDefaults(t *testing.T) {
	d, bin := setupModelsRepo(t, "")
	cmd := exec.Command(bin, "hook", "orchestration")
	cmd.Dir = d
	out, _ := cmd.CombinedOutput()
	s := string(out)
	if !strings.Contains(s, "big-thinker (model: reasoning)") {
		t.Errorf("AC3: big-thinker default; got:\n%s", s)
	}
	if !strings.Contains(s, "feature-specialist (model: balanced)") {
		t.Errorf("AC3: feature-specialist default; got:\n%s", s)
	}
}

// Edge: out-of-band roles must not appear in the directive.
func TestOrchestrationHook_OutOfBandRolesAbsent(t *testing.T) {
	d, bin := setupModelsRepo(t, "")
	cmd := exec.Command(bin, "hook", "orchestration")
	cmd.Dir = d
	out, _ := cmd.CombinedOutput()
	s := string(out)
	for _, oob := range []string{"gatekeeper", "production-readiness", "edge-case-tester", "merge-steward"} {
		if strings.Contains(s, oob+" (model:") {
			t.Errorf("out-of-band role %q should not be annotated; got:\n%s", oob, s)
		}
	}
}
