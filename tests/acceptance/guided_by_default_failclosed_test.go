// Acceptance: specs/guided-by-default.feature
package acceptance_test

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

// brokenTOML is a centinela.toml that pins strict and cannot be parsed. Every
// profile surface must answer strict for it — the pin is beside the point; an
// unreadable config makes the profile unknowable and unknowable means strict.
const brokenTOML = "[workflow]\nenforcement_profile = \"strict\"\nuse_worktrees = tru\n"

// Scenario: An explicit global profile still outranks the new default
//
// Round 1 closed this direction in `hook autostart`; it was still open in
// `status`, one file away. Every REPORTING surface now routes through the shared
// resolver, so none of them can answer looser than the enforcing surfaces.
func TestGBD_StatusFailsClosedOnUnreadableConfig(t *testing.T) {
	bin := buildCent(t)
	dir := t.TempDir()
	gbdColdTree(t, dir)
	mustWrite(t, filepath.Join(dir, "centinela.toml"), brokenTOML)
	mustWrite(t, filepath.Join(dir, ".workflow", "f.json"),
		gbdPinnedWorkflowJSON("totally-unknown-model"))

	out, code := runCent(t, bin, dir, "status", "f")
	if code != 0 {
		t.Fatalf("status must still render, exit %d: %s", code, out)
	}
	if strings.Contains(out, "guided") {
		t.Fatalf("status must not report the loosest profile on a broken config: %s", out)
	}
	if !strings.Contains(out, "strict") {
		t.Fatalf("status must fail closed to strict: %s", out)
	}
}

// Scenario: An explicit global profile still outranks the new default
//
// verdict is the machine payload. On the same broken config it must agree with
// status, and `start` must refuse outright — no surface may be looser than
// another about the same tree.
func TestGBD_VerdictAndStartAgreeOnUnreadableConfig(t *testing.T) {
	bin := buildCent(t)
	dir := t.TempDir()
	gbdColdTree(t, dir)
	mustWrite(t, filepath.Join(dir, "centinela.toml"), brokenTOML)
	mustWrite(t, filepath.Join(dir, ".workflow", "f.json"),
		gbdPinnedWorkflowJSON("totally-unknown-model"))

	verdictOut, _ := runCent(t, bin, dir, "verdict", "f")
	var packet map[string]any
	if err := json.Unmarshal([]byte(verdictOut), &packet); err == nil {
		if got := gbdVerdictProfile(t, packet); got != "strict" {
			t.Fatalf("verdict on a broken config = %q, want strict", got)
		}
	}
	out, code := runCent(t, bin, dir, "start", "setup")
	if code == 0 {
		t.Fatalf("start must refuse an unreadable config outright: %s", out)
	}
	if !strings.Contains(out, "centinela.toml") {
		t.Fatalf("the refusal must name the file: %s", out)
	}
}
