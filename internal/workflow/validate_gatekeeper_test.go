package workflow

import (
	"os"
	"strings"
	"testing"
)

const groundedReport = "**Status:** SAFE\n\n```json centinela:verification\n" +
	`{"revision":"abc","treeDigest":"sha256:d","commands":[{"argv":["centinela","validate"],"exitCode":0}]}` +
	"\n```\n"

// seedGate chdirs into a temp repo carrying a workflow with the given pinned
// contract and the given gatekeeper report body.
func seedGate(t *testing.T, contract, report string) {
	t.Helper()
	t.Chdir(t.TempDir())
	if err := os.MkdirAll(WorkflowDir, 0o755); err != nil {
		t.Fatal(err)
	}
	wf := New("f")
	wf.ValidateContract = contract
	if err := Save(wf); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(GatekeeperReportPath("f"), []byte(report), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestValidateGatekeeperLegacyAcceptsOldFormat(t *testing.T) {
	seedGate(t, "", "### Gatekeeper Report\n**Status:** SAFE\n")
	if err := validateGatekeeper("f"); err != nil {
		t.Fatalf("legacy contract must keep the existence-only gate: %v", err)
	}
}

func TestValidateGatekeeperAdversarialRejectsOldFormat(t *testing.T) {
	seedGate(t, ValidateContractAdversarial, "### Gatekeeper Report\n**Status:** SAFE\n")
	err := validateGatekeeper("f")
	if err == nil || !strings.Contains(err.Error(), "no commands-run record") {
		t.Fatalf("want no-commands-run-record block, got %v", err)
	}
}

func TestValidateGatekeeperAdversarialAcceptsGrounded(t *testing.T) {
	seedGate(t, ValidateContractAdversarial, groundedReport)
	if err := validateGatekeeper("f"); err != nil {
		t.Fatalf("grounded report rejected: %v", err)
	}
}

func TestValidateGatekeeperCriticalBlocksWithReason(t *testing.T) {
	seedGate(t, ValidateContractAdversarial, "**Status:** CRITICAL\n\n#### Findings\n- tokens leak\n")
	err := validateGatekeeper("f")
	if err == nil || !strings.Contains(err.Error(), "tokens leak") {
		t.Fatalf("want echoed finding, got %v", err)
	}
}

func TestValidateGatekeeperMissingReport(t *testing.T) {
	seedGate(t, ValidateContractAdversarial, groundedReport)
	os.Remove(GatekeeperReportPath("f")) //nolint:errcheck
	if err := validateGatekeeper("f"); err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("want not-found error, got %v", err)
	}
}

func TestUsesAdversarialVerifier(t *testing.T) {
	if (*Workflow)(nil).UsesAdversarialVerifier() {
		t.Fatal("nil workflow must be legacy")
	}
	if (&Workflow{}).UsesAdversarialVerifier() {
		t.Fatal("unpinned workflow must be legacy")
	}
	if !New("f").UsesAdversarialVerifier() {
		t.Fatal("NewWithOrder must pin the adversarial contract")
	}
}

func TestFeatureUsesAdversarialVerifierMissingState(t *testing.T) {
	t.Chdir(t.TempDir())
	if featureUsesAdversarialVerifier("nope") {
		t.Fatal("an unreadable workflow must be treated as legacy")
	}
}
