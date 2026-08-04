// Acceptance: specs/durable-workflow-state.feature
package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Scenario: A newly started workflow is stamped with the current schema version
func TestAccStartStampsTheSchemaVersion(t *testing.T) {
	bin := dmrBuildBin(t)
	dir := dmrRepo(t, dmrDynamicTOML)

	dmrOK(t, bin, dir, "start", "beta")
	got := dmrState(t, dir, "beta")
	if !strings.Contains(got, `"schemaVersion": 1`) {
		t.Fatalf("a newly started workflow must be stamped:\n%s", got)
	}
	if idx := strings.Index(got, `"schemaVersion"`); idx > 5 {
		t.Fatalf("schemaVersion must be the first key, found at %d:\n%s", idx, got)
	}
}

// Scenario: A workflow that was never loaded saves without a conflict check
//
// `start` builds its workflow rather than loading one, so it carries no digest
// and never hits the staleness check. What keeps that exemption safe is the
// already-exists fence in the command itself — pinned here, since nothing in
// internal/workflow enforces it.
func TestAccStartNeverLoadsAndIsFencedByExistence(t *testing.T) {
	bin := dmrBuildBin(t)
	dir := dmrRepo(t, dmrDynamicTOML)

	dmrOK(t, bin, dir, "start", "zeta")
	first := dmrState(t, dir, "zeta")
	if !strings.Contains(first, `"currentStep": "plan"`) {
		t.Fatalf("a new workflow must record its first step:\n%s", first)
	}

	out := dmrRefused(t, bin, dir, "start", "zeta")
	if !strings.Contains(out, "already exists") {
		t.Fatalf("a second start must be refused by the existence fence, got:\n%s", out)
	}
	if dmrState(t, dir, "zeta") != first {
		t.Fatal("the refused start rewrote the existing state file")
	}
}

// Scenario: An abandoned temporary file is not mistaken for orphaned evidence
func TestAccDoctorIgnoresAnAbandonedWorkflowTemp(t *testing.T) {
	bin := dmrBuildBin(t)
	dir := dmrRepo(t, dmrDynamicTOML)
	dmrWrite(t, dir, "f", dmrWorkflowJSON("f", "code", ""))
	tmp := filepath.Join(dir, ".workflow", ".f.json.tmp-1234567890")
	if err := os.WriteFile(tmp, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}

	out, _ := runCent(t, bin, dir, "doctor")
	if !strings.Contains(out, "no orphaned evidence temp files") {
		t.Fatalf("the evidence check must stay green, got:\n%s", out)
	}
	if strings.Contains(out, ".f.json.tmp-") {
		t.Fatalf("doctor must not report the workflow temp, got:\n%s", out)
	}
	if _, err := os.Stat(tmp); err != nil {
		t.Fatalf("doctor must not touch it: %v", err)
	}
}
