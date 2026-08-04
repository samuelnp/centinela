package integration_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// hcpdBuild compiles the binary under test into a temp dir. Local build only —
// no network, no shared install path.
func hcpdBuild(t *testing.T) string {
	t.Helper()
	wd, _ := os.Getwd()
	repo := filepath.Clean(filepath.Join(wd, "..", ".."))
	bin := filepath.Join(t.TempDir(), "centinela")
	c := exec.Command("go", "build", "-o", bin, "./cmd/centinela")
	c.Dir = repo
	if out, err := c.CombinedOutput(); err != nil {
		t.Fatalf("build centinela: %v\n%s", err, out)
	}
	return bin
}

// hcpdRepo seeds a throwaway project directory. Never the real repo: hook
// context reads .workflow/ from its working directory.
func hcpdRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".workflow"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "PROJECT.md"), []byte("Project Stage: existing\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	return dir
}

func hcpdWorkflow(t *testing.T, dir, feature, step string, order []string) {
	t.Helper()
	quoted := make([]string, len(order))
	for i, s := range order {
		quoted[i] = `"` + s + `"`
	}
	body := `{"feature":"` + feature + `","currentStep":"` + step + `",` +
		`"stepOrder":[` + strings.Join(quoted, ",") + `],"steps":{}}`
	if err := os.WriteFile(filepath.Join(dir, ".workflow", feature+".json"), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

func hcpdHookContext(t *testing.T, bin, dir string) string {
	t.Helper()
	c := exec.Command(bin, "hook", "context")
	c.Dir = dir
	c.Stdin = strings.NewReader(`{"session_id":"hcpd-1"}`)
	out, err := c.CombinedOutput()
	if err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			t.Fatalf("run hook context: %v", err)
		}
	}
	return string(out)
}

func hcpdPadded(s string) []string {
	var out []string
	for _, line := range strings.Split(s, "\n") {
		if line != strings.TrimRight(line, " \t") {
			out = append(out, line)
		}
	}
	return out
}

var hcpdCanonical = []string{"plan", "code", "tests", "validate", "docs"}

// TestHookContextEmitsNoTrailingWhitespace drives the real binary end to end
// with a mixed-archetype, multi-workflow fixture — the shape that maximises
// lipgloss padding, since sibling lines differ most in width.
func TestHookContextEmitsNoTrailingWhitespace(t *testing.T) {
	bin := hcpdBuild(t)
	dir := hcpdRepo(t)
	hcpdWorkflow(t, dir, "a", "plan", hcpdCanonical)
	hcpdWorkflow(t, dir, "a-considerably-longer-feature-slug", "docs", hcpdCanonical)
	hcpdWorkflow(t, dir, "urgent-fix", "code", []string{"code", "tests", "validate"})
	hcpdWorkflow(t, dir, "quick-check", "code", []string{"plan", "code"})

	out := hcpdHookContext(t, bin, dir)
	if bad := hcpdPadded(out); len(bad) > 0 {
		t.Fatalf("hook context emitted %d padded line(s), first %q\nfull output:\n%s", len(bad), bad[0], out)
	}
	for _, want := range []string{"ACTIVE WORKFLOWS", "urgent-fix", "quick-check", "FEATURE BRIEF REQUIRED"} {
		if !strings.Contains(out, want) {
			t.Errorf("hook context lost %q:\n%s", want, out)
		}
	}
}

// TestHookContextKeepsPerArchetypeLadder pins that the trimmed panel still
// distinguishes a spike (no validate step, ungated by construction) from a
// hotfix (no plan step) — the governance signal the plan refused to cut.
func TestHookContextKeepsPerArchetypeLadder(t *testing.T) {
	bin := hcpdBuild(t)
	dir := hcpdRepo(t)
	hcpdWorkflow(t, dir, "sample-feature", "plan", hcpdCanonical)
	hcpdWorkflow(t, dir, "quick-check", "code", []string{"plan", "code"})
	hcpdWorkflow(t, dir, "urgent-fix", "code", []string{"code", "tests", "validate"})

	out := hcpdHookContext(t, bin, dir)
	ladders := map[string]string{}
	lines := strings.Split(out, "\n")
	for i, line := range lines {
		for _, f := range []string{"sample-feature", "quick-check", "urgent-fix"} {
			if strings.Contains(line, f) && i+1 < len(lines) && strings.Contains(lines[i+1], "·") {
				ladders[f] = lines[i+1]
			}
		}
	}
	if len(ladders) != 3 {
		t.Fatalf("expected a step ladder per workflow, got %d:\n%s", len(ladders), out)
	}
	if strings.Contains(ladders["quick-check"], "validate") {
		t.Errorf("spike ladder must not show a validate step: %q", ladders["quick-check"])
	}
	if strings.Contains(ladders["urgent-fix"], "plan") {
		t.Errorf("hotfix ladder must not show a plan step: %q", ladders["urgent-fix"])
	}
	if !strings.Contains(ladders["sample-feature"], "validate") {
		t.Errorf("canonical ladder lost its validate step: %q", ladders["sample-feature"])
	}
}

// TestHookContextZeroWorkflowsStillDirects covers the empty-state path, which
// renders through renderSystemLine (never padded) and must stay intact.
func TestHookContextZeroWorkflowsStillDirects(t *testing.T) {
	out := hcpdHookContext(t, hcpdBuild(t), hcpdRepo(t))
	if bad := hcpdPadded(out); len(bad) > 0 {
		t.Errorf("empty-state hook context padded a line: %q", bad[0])
	}
	if !strings.Contains(out, "no active workflow") {
		t.Errorf("empty-state directive lost:\n%s", out)
	}
}
