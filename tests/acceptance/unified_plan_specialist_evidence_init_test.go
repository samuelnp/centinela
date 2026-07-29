// Acceptance: specs/unified-plan-specialist.feature
package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Scenario: evidence init for a retired legacy role errors on a pinned workflow
func TestUPS_EvidenceInitRefusesRetiredRoleOnPinnedWorkflow(t *testing.T) {
	bin := upsBuildBin(t)
	dir := upsExistingRepo(t)
	upsWriteWorkflow(t, dir, "new-widget", "planner-v1")

	out, code := runCent(t, bin, dir, "evidence", "init", "new-widget", "big-thinker")
	if code == 0 {
		t.Fatalf("evidence init for a retired role must fail on a pinned workflow: %s", out)
	}
	if !strings.Contains(out, "retired") || !strings.Contains(out, "planner") {
		t.Fatalf("error must say the role is retired and point at planner: %s", out)
	}
	if _, err := os.Stat(filepath.Join(dir, ".workflow", "new-widget-big-thinker.json")); !os.IsNotExist(err) {
		t.Fatal("a refused init must not write a stub")
	}
}

// Scenario: evidence init for a retired legacy role errors with no workflow state at all
func TestUPS_EvidenceInitRefusesRetiredRoleWithNoWorkflow(t *testing.T) {
	bin := upsBuildBin(t)
	dir := upsExistingRepo(t) // .workflow/ exists, but no ghost-feature.json

	out, code := runCent(t, bin, dir, "evidence", "init", "ghost-feature", "feature-specialist")
	if code == 0 {
		t.Fatalf("evidence init for a retired role must fail with no workflow state: %s", out)
	}
	if !strings.Contains(out, "retired") || !strings.Contains(out, "planner") {
		t.Fatalf("error must say the role is retired and point at planner: %s", out)
	}
}
