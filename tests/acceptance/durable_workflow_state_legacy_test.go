// Acceptance: specs/durable-workflow-state.feature
package acceptance_test

import (
	"os/exec"
	"strings"
	"testing"
)

// dwsPrewrite drives the real prewrite hook over a pending write to rel,
// exactly as the harness does: the payload arrives on stdin.
func dwsPrewrite(t *testing.T, bin, dir, rel string) (string, int) {
	t.Helper()
	c := exec.Command(bin, "hook", "prewrite")
	c.Dir = dir
	c.Stdin = strings.NewReader(`{"tool_input":{"file_path":"` + dir + "/" + rel + `"}}`)
	out, err := c.CombinedOutput()
	code := 0
	if ee, ok := err.(*exec.ExitError); ok {
		code = ee.ExitCode()
	} else if err != nil {
		t.Fatalf("run hook prewrite: %v", err)
	}
	return string(out), code
}

// Scenario: A versionless legacy workflow file loads unchanged
//
// Also covers "A newly started workflow is stamped with the current schema
// version" from the other direction: the first save-bearing command stamps the
// legacy file, and the key lands first so `head -3` of any state file shows it.
func TestAccLegacyVersionlessFileLoadsAndIsStamped(t *testing.T) {
	bin := dmrBuildBin(t)
	dir := dmrRepo(t, dmrDynamicTOML)
	legacy := dmrLegacyWorkflowJSON("legacy", "code", "")
	dmrWrite(t, dir, "legacy", legacy)
	if strings.Contains(legacy, "schemaVersion") {
		t.Fatal("precondition: the fixture must predate schema versions")
	}

	out := dmrOK(t, bin, dir, "status", "legacy")
	if !strings.Contains(out, "legacy") {
		t.Fatalf("a versionless file must load: %s", out)
	}
	if got := dmrState(t, dir, "legacy"); got != legacy {
		t.Fatalf("a read-only command must not rewrite the file:\n%s", got)
	}

	dmrOK(t, bin, dir, "route", "set", "legacy", "senior-engineer", "balanced",
		"--reason", "config-only change")
	stamped := dmrState(t, dir, "legacy")
	if !strings.Contains(stamped, `"schemaVersion": 1`) {
		t.Fatalf("the first save must stamp the current version:\n%s", stamped)
	}
	if idx := strings.Index(stamped, `"schemaVersion"`); idx > 5 {
		t.Fatalf("schemaVersion must be the first key, found at %d:\n%s", idx, stamped)
	}
	if !strings.Contains(stamped, "modelRoutes") {
		t.Fatalf("the recorded route must survive the stamp:\n%s", stamped)
	}
}
