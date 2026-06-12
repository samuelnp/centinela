package integration_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

// End-to-end: `centinela start --archetype <name>` resolves and persists the
// archetype's step order into the .workflow state — the order flows from the
// real start path, not a test shortcut.
func TestStartArchetypePersistsOrder(t *testing.T) {
	o, _ := os.Getwd()
	repo := filepath.Clean(filepath.Join(o, "..", ".."))
	bin := filepath.Join(t.TempDir(), "centinela-arch-int")
	build := exec.Command("go", "build", "-o", bin, "./cmd/centinela")
	build.Dir = repo
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}

	cases := map[string][]string{
		"hotfix": {"code", "tests", "validate"},
		"spike":  {"plan", "code"},
	}
	for arch, want := range cases {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "PROJECT.md"), []byte("Project Stage: existing\n"), 0644) //nolint:errcheck
		start := exec.Command(bin, "start", "feat", "--archetype", arch)
		start.Dir = dir
		if out, err := start.CombinedOutput(); err != nil {
			t.Fatalf("%s start failed: %v\n%s", arch, err, out)
		}
		wf := loadFromDir(t, dir, "feat")
		if wf.Archetype != arch {
			t.Fatalf("%s: persisted archetype = %q", arch, wf.Archetype)
		}
		if !slices.Equal(wf.StepOrder, want) {
			t.Fatalf("%s: order = %v, want %v", arch, wf.StepOrder, want)
		}
		if slices.Contains(want, "validate") != slices.Contains(wf.StepOrder, "validate") {
			t.Fatalf("%s: validate-presence mismatch in persisted order %v", arch, wf.StepOrder)
		}
	}
}

func loadFromDir(t *testing.T, dir, feature string) *workflow.Workflow {
	t.Helper()
	t.Chdir(dir)
	wf, err := workflow.Load(feature)
	if err != nil {
		t.Fatalf("load workflow: %v", err)
	}
	return wf
}
