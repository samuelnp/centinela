// Acceptance: specs/guided-by-default.feature
package acceptance_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// runCentStdin runs the binary in dir with a hook payload on stdin, returning
// combined output and the exit code.
func runCentStdin(t *testing.T, bin, dir, stdin string, args ...string) (string, int) {
	t.Helper()
	c := exec.Command(bin, args...)
	c.Dir = dir
	c.Stdin = strings.NewReader(stdin)
	out, err := c.CombinedOutput()
	code := 0
	if ee, ok := err.(*exec.ExitError); ok {
		code = ee.ExitCode()
	} else if err != nil {
		t.Fatalf("run %v: %v", args, err)
	}
	return string(out), code
}

// gbdAutostartTree lays a startable existing-project tree with the given config.
func gbdAutostartTree(t *testing.T, toml string) string {
	t.Helper()
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "PROJECT.md"), "Project Stage: existing\n")
	if toml != "" {
		mustWrite(t, filepath.Join(dir, "centinela.toml"), toml)
	}
	return dir
}

// gbdWorkflowStates returns every workflow state file the hook wrote, excluding
// the roadmap artifacts that share the directory.
func gbdWorkflowStates(t *testing.T, dir string) []string {
	t.Helper()
	paths, _ := filepath.Glob(filepath.Join(dir, ".workflow", "*.json"))
	var out []string
	for _, p := range paths {
		var m map[string]any
		body, err := os.ReadFile(p)
		if err != nil || json.Unmarshal(body, &m) != nil {
			continue
		}
		if _, ok := m["currentStep"]; ok {
			out = append(out, string(body))
		}
	}
	return out
}

// Scenario: An explicit global profile still outranks the new default
//
// The autostart hook resolves the profile from centinela.toml. When that file
// cannot be READ, the profile is unknowable — and because the contract pin is
// written once and for life, guessing guided here can never be undone by fixing
// the typo. So it must fail CLOSED, exactly as the setup hook does.
func TestGBD_AutostartFailsClosedOnUnparseableConfig(t *testing.T) {
	bin := buildCent(t)
	dir := gbdAutostartTree(t, "[workflow]\nenforcement_profile = \"strict\"\nuse_worktrees = tru\n")
	out, code := runCentStdin(t, bin, dir,
		`{"prompt":"please add release diagnostics"}`, "hook", "autostart")
	if code != 0 {
		t.Fatalf("a hook must never break the session, exit %d: %s", code, out)
	}
	if !strings.Contains(out, "autostart declined") || !strings.Contains(out, "centinela.toml") {
		t.Fatalf("output must decline and name the parse failure: %s", out)
	}
	if states := gbdWorkflowStates(t, dir); len(states) != 0 {
		t.Fatalf("no workflow may be pinned on an unreadable config: %v", states)
	}
}

// Scenario: An explicit global profile still outranks the new default
//
// The same config, minus the typo, must yield STRICT and never the guided tail.
func TestGBD_AutostartHonoursStrictWhenTheConfigParses(t *testing.T) {
	bin := buildCent(t)
	dir := gbdAutostartTree(t, "[workflow]\nenforcement_profile = \"strict\"\n")
	if out, code := runCentStdin(t, bin, dir,
		`{"prompt":"please add release diagnostics"}`, "hook", "autostart"); code != 0 {
		t.Fatalf("autostart exit %d: %s", code, out)
	}
	states := gbdWorkflowStates(t, dir)
	if len(states) != 1 {
		t.Fatalf("expected exactly one autostarted workflow, got %v", states)
	}
	if !strings.Contains(states[0], "strict-subagents-v1") {
		t.Fatalf("strict must still require the orchestration evidence bundle: %s", states[0])
	}
}
