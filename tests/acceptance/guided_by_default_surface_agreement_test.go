// Acceptance: specs/guided-by-default.feature
package acceptance_test

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

// gbdPinnedWorkflowJSON is a workflow born AFTER the flip (it carries the
// profile contract) that also declares a driver model Centinela has no
// capability class for — the exact state on which the human and machine
// surfaces used to disagree.
func gbdPinnedWorkflowJSON(driver string) string {
	return `{"feature":"f","currentStep":"plan",` +
		`"stepOrder":["plan","code","tests","validate","docs"],` +
		`"steps":{"plan":{"status":"in-progress","completedAt":null}},` +
		`"driverModel":"` + driver + `","profileContract":"guided-default-v1"}`
}

// gbdVerdictProfile pulls the resolved profile out of the verdict packet's run
// block — the machine-readable value MCP clients consume.
func gbdVerdictProfile(t *testing.T, packet map[string]any) string {
	t.Helper()
	run, ok := packet["run"].(map[string]any)
	if !ok {
		t.Fatalf("verdict packet has no run block: %v", packet)
	}
	profile, _ := run["profile"].(string)
	return profile
}

// Scenario: A driver model's capability class still outranks the new default
//
// `centinela status` (human) and `centinela verdict` (machine, also served over
// MCP) read the SAME state file through two different resolvers. They must never
// report different profiles: a governance payload that contradicts the terminal
// is worse than either answer alone.
func TestGBD_StatusAndVerdictAgreeOnUnmappedDriver(t *testing.T) {
	bin := buildCent(t)
	dir := t.TempDir()
	gbdColdTree(t, dir)
	mustWrite(t, filepath.Join(dir, ".workflow", "f.json"),
		gbdPinnedWorkflowJSON("totally-unknown-model"))

	statusOut, code := runCent(t, bin, dir, "status", "f")
	if code != 0 {
		t.Fatalf("status exit %d: %s", code, statusOut)
	}
	verdictOut, _ := runCent(t, bin, dir, "verdict", "f")
	var packet map[string]any
	if err := json.Unmarshal([]byte(verdictOut), &packet); err != nil {
		t.Fatalf("verdict must emit JSON: %v\n%s", err, verdictOut)
	}
	machine := gbdVerdictProfile(t, packet)
	if machine != "strict" {
		t.Fatalf("a declared-but-unmapped driver must keep strict, verdict says %q", machine)
	}
	if !strings.Contains(statusOut, machine) {
		t.Fatalf("status and verdict disagree: verdict %q, status:\n%s", machine, statusOut)
	}
	if !strings.Contains(statusOut, "no capability") {
		t.Fatalf("status must still explain WHY it is strict: %s", statusOut)
	}
}

// Scenario: A capability-derived guided profile is distinguishable from the default
//
// The ✅ counterweight, so agreement is not achieved by answering "strict" to
// everything.
func TestGBD_StatusAndVerdictAgreeOnCapableDriver(t *testing.T) {
	bin := buildCent(t)
	dir := t.TempDir()
	gbdColdTree(t, dir)
	mustWrite(t, filepath.Join(dir, ".workflow", "f.json"), gbdPinnedWorkflowJSON("sonnet"))

	statusOut, _ := runCent(t, bin, dir, "status", "f")
	verdictOut, _ := runCent(t, bin, dir, "verdict", "f")
	var packet map[string]any
	if err := json.Unmarshal([]byte(verdictOut), &packet); err != nil {
		t.Fatalf("verdict must emit JSON: %v\n%s", err, verdictOut)
	}
	if machine := gbdVerdictProfile(t, packet); machine != "guided" {
		t.Fatalf("a capable driver must resolve guided, verdict says %q", machine)
	}
	if !strings.Contains(statusOut, "guided") {
		t.Fatalf("status must agree with the verdict: %s", statusOut)
	}
	if !strings.Contains(statusOut, "driver: sonnet") {
		t.Fatalf("status must name the driver, not the shipped default: %s", statusOut)
	}
}
