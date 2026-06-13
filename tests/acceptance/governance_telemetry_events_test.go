package acceptance_test

import (
	"testing"

	"github.com/samuelnp/centinela/internal/telemetry"
)

// Acceptance: specs/governance-telemetry.feature

// Scenario: A failing gate during validate appends a gate-failure event
func TestGT_GateFailure(t *testing.T) {
	gtChdir(t)
	telemetry.RecordGateFailure(gtCfg(true), "G1: File Size", "big.go exceeds 100 lines")
	evs := gtEvents(t)
	if len(evs) != 1 || evs[0].Type != telemetry.TypeGateFailure ||
		evs[0].Gate != "G1: File Size" || evs[0].Message == "" {
		t.Fatalf("gate-failure event wrong: %+v", evs)
	}
}

// Scenario: Each failing gate appends its own gate-failure event
func TestGT_GateFailurePerGate(t *testing.T) {
	gtChdir(t)
	cfg := gtCfg(true)
	telemetry.RecordGateFailure(cfg, "G1: File Size", "m1")
	telemetry.RecordGateFailure(cfg, "import_graph", "m2")
	evs := gtEvents(t)
	if len(evs) != 2 || evs[0].Gate != "G1: File Size" || evs[1].Gate != "import_graph" {
		t.Fatalf("want one event per failing gate, got %+v", evs)
	}
}

// Scenario: A failed claim verification appends a verify-rejection event with the failing checks
func TestGT_VerifyRejection(t *testing.T) {
	gtChdir(t)
	checks := []telemetry.CheckRef{{Claim: "coverage", Role: "qa-senior", Status: "FAIL", Detail: "92% < 95%"}}
	telemetry.RecordVerifyRejection(gtCfg(true), "feat", "validate", checks)
	evs := gtEvents(t)
	if len(evs) != 1 || evs[0].Type != telemetry.TypeVerifyRejection || len(evs[0].Checks) != 1 ||
		evs[0].Checks[0].Claim != "coverage" || evs[0].Checks[0].Status != "FAIL" {
		t.Fatalf("verify-rejection event wrong: %+v", evs)
	}
}

// Scenario: An advance aborted by validate gates appends complete-rejected with reason gates
func TestGT_CompleteRejectedGates(t *testing.T) {
	gtChdir(t)
	telemetry.RecordCompleteRejected(gtCfg(true), "feat", "validate", "gates")
	evs := gtEvents(t)
	if len(evs) != 1 || evs[0].Type != telemetry.TypeCompleteRejected ||
		evs[0].Reason != "gates" || evs[0].Feature != "feat" || evs[0].Step != "validate" {
		t.Fatalf("complete-rejected(gates) wrong: %+v", evs)
	}
}

// Scenario: An advance aborted by verification appends complete-rejected with reason verify
func TestGT_CompleteRejectedVerify(t *testing.T) {
	gtChdir(t)
	telemetry.RecordCompleteRejected(gtCfg(true), "feat", "validate", "verify")
	evs := gtEvents(t)
	if len(evs) != 1 || evs[0].Reason != "verify" {
		t.Fatalf("complete-rejected(verify) wrong: %+v", evs)
	}
}

// Scenario: A successful advance appends a step-advanced event carrying the just-completed step
func TestGT_StepAdvanced(t *testing.T) {
	gtChdir(t)
	telemetry.RecordStepAdvanced(gtCfg(true), "feat", "plan")
	evs := gtEvents(t)
	if len(evs) != 1 || evs[0].Type != telemetry.TypeStepAdvanced ||
		evs[0].Feature != "feat" || evs[0].Step != "plan" {
		t.Fatalf("step-advanced wrong: %+v", evs)
	}
}
