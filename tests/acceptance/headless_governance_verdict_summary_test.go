package acceptance_test

import (
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/gates"
	"github.com/samuelnp/centinela/internal/verdict"
	"github.com/samuelnp/centinela/internal/verify"
)

// Acceptance: specs/headless-governance.feature

func hgAssemble(g []gates.Result, v verify.VerificationResult, e []verdict.EvidLine) *verdict.Packet {
	return verdict.AssembleVerdict("headless-governance", &config.Config{}, hgWf(), hgDeps(g, v, e))
}

// Scenario: Verdict packet pass emits exit code zero and JSON to stdout
func TestHG_VerdictPass(t *testing.T) {
	t.Setenv("CENTINELA_HEADLESS", "")
	pkt := hgAssemble([]gates.Result{hgPassGate()}, hgVerify(), nil)
	if pkt.Summary.Verdict != "pass" || pkt.Summary.ExitCode != 0 {
		t.Fatalf("want pass/0, got %s/%d", pkt.Summary.Verdict, pkt.Summary.ExitCode)
	}
}

// Scenario: Verdict packet fail still emits JSON to stdout with exit code one
func TestHG_VerdictGateFail(t *testing.T) {
	t.Setenv("CENTINELA_HEADLESS", "")
	fail := gates.Result{Name: "G1: File Size", Status: gates.Fail, Message: "too long"}
	pkt := hgAssemble([]gates.Result{fail}, hgVerify(), nil)
	if pkt.Summary.Verdict != "fail" || pkt.Summary.ExitCode != 1 {
		t.Fatalf("want fail/1, got %s/%d", pkt.Summary.Verdict, pkt.Summary.ExitCode)
	}
}

// Scenario: A verify failure alone produces a fail verdict
func TestHG_VerifyFailAlone(t *testing.T) {
	t.Setenv("CENTINELA_HEADLESS", "")
	v := hgVerify(verify.Check{Role: "qa-senior", Status: verify.StatusFail})
	pkt := hgAssemble([]gates.Result{hgPassGate()}, v, nil)
	if pkt.Summary.Verdict != "fail" || pkt.Summary.ExitCode != 1 {
		t.Fatalf("verify fail alone must fail, got %s/%d", pkt.Summary.Verdict, pkt.Summary.ExitCode)
	}
}

// Scenario: Warnings are reported but do not fail the verdict
func TestHG_WarningsDontFail(t *testing.T) {
	t.Setenv("CENTINELA_HEADLESS", "")
	v := hgVerify(verify.Check{Role: "qa-senior", Status: verify.StatusWarn})
	pkt := hgAssemble([]gates.Result{hgPassGate()}, v, nil)
	if pkt.Summary.Verdict != "pass" || pkt.Summary.Verify.Warn != 1 {
		t.Fatalf("warn must report but not fail: %s warn=%d", pkt.Summary.Verdict, pkt.Summary.Verify.Warn)
	}
}

// Scenario: Packet schema field is the versioned identifier
func TestHG_SchemaField(t *testing.T) {
	t.Setenv("CENTINELA_HEADLESS", "")
	if pkt := hgAssemble([]gates.Result{hgPassGate()}, hgVerify(), nil); pkt.Schema != "centinela.verdict/v1" {
		t.Fatalf("schema = %q", pkt.Schema)
	}
}

// Scenario: Gate statuses are lowercased and verify statuses stay uppercase
func TestHG_StatusCasing(t *testing.T) {
	t.Setenv("CENTINELA_HEADLESS", "")
	v := hgVerify(verify.Check{Role: "qa-senior", Status: verify.StatusPass})
	pkt := hgAssemble([]gates.Result{hgPassGate()}, v, nil)
	if pkt.Gates[0].Status != "pass" || pkt.Verify[0].Status != "PASS" {
		t.Fatalf("casing wrong: gate=%q verify=%q", pkt.Gates[0].Status, pkt.Verify[0].Status)
	}
}
