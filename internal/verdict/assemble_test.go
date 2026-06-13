package verdict

import (
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/gates"
	"github.com/samuelnp/centinela/internal/verify"
	"github.com/samuelnp/centinela/internal/workflow"
)

func wfValidate() *workflow.Workflow {
	return &workflow.Workflow{Feature: "headless-governance", CurrentStep: "validate", DriverModel: "claude-opus"}
}

// All gates pass and verify reports no failures → pass / exit 0.
func TestAssembleVerdict_Pass(t *testing.T) {
	deps := fakeDeps([]gates.Result{passGate()}, vr(verify.Check{Role: "qa-senior", Status: verify.StatusPass}), nil)
	pkt := AssembleVerdict("headless-governance", &config.Config{}, wfValidate(), deps)
	if pkt.Schema != "centinela.verdict/v1" {
		t.Fatalf("schema = %q", pkt.Schema)
	}
	if pkt.Summary.Verdict != "pass" || pkt.Summary.ExitCode != 0 {
		t.Fatalf("want pass/0, got %s/%d", pkt.Summary.Verdict, pkt.Summary.ExitCode)
	}
	if pkt.Summary.Gates.Pass != 1 || pkt.Summary.Verify.Pass != 1 {
		t.Fatalf("counts = %+v / %+v", pkt.Summary.Gates, pkt.Summary.Verify)
	}
}

// A gate Fail produces a fail verdict / exit 1.
func TestAssembleVerdict_GateFail(t *testing.T) {
	deps := fakeDeps([]gates.Result{failGate()}, vr(), nil)
	pkt := AssembleVerdict("headless-governance", &config.Config{}, wfValidate(), deps)
	if pkt.Summary.Verdict != "fail" || pkt.Summary.ExitCode != 1 {
		t.Fatalf("gate fail must be fail/1, got %s/%d", pkt.Summary.Verdict, pkt.Summary.ExitCode)
	}
	if pkt.Summary.Gates.Fail != 1 {
		t.Fatalf("gate fail count = %d", pkt.Summary.Gates.Fail)
	}
}

// A verify failure alone (all gates pass) still fails the verdict.
func TestAssembleVerdict_VerifyFail(t *testing.T) {
	deps := fakeDeps([]gates.Result{passGate()}, vr(verify.Check{Role: "qa-senior", Status: verify.StatusFail}), nil)
	pkt := AssembleVerdict("headless-governance", &config.Config{}, wfValidate(), deps)
	if pkt.Summary.Verdict != "fail" || pkt.Summary.ExitCode != 1 {
		t.Fatalf("verify fail must be fail/1, got %s/%d", pkt.Summary.Verdict, pkt.Summary.ExitCode)
	}
}

// A WARN check is tallied but does not fail the verdict.
func TestAssembleVerdict_WarnsDontFail(t *testing.T) {
	deps := fakeDeps([]gates.Result{passGate()}, vr(verify.Check{Role: "qa-senior", Status: verify.StatusWarn}), nil)
	pkt := AssembleVerdict("headless-governance", &config.Config{}, wfValidate(), deps)
	if pkt.Summary.Verdict != "pass" || pkt.Summary.ExitCode != 0 {
		t.Fatalf("warn must stay pass/0, got %s/%d", pkt.Summary.Verdict, pkt.Summary.ExitCode)
	}
	if pkt.Summary.Verify.Warn != 1 {
		t.Fatalf("warn count = %d", pkt.Summary.Verify.Warn)
	}
}

// Run info snapshots feature/step/profile/archetype/driverModel/headless/now.
func TestAssembleVerdict_RunInfoProvenance(t *testing.T) {
	t.Setenv("CENTINELA_HEADLESS", "1")
	deps := fakeDeps([]gates.Result{passGate()}, vr(), nil)
	pkt := AssembleVerdict("headless-governance", &config.Config{}, wfValidate(), deps)
	r := pkt.Run
	if r.Feature != "headless-governance" || r.Step != "validate" {
		t.Fatalf("feature/step = %q/%q", r.Feature, r.Step)
	}
	if r.Profile != "strict" || r.Archetype != "canonical" {
		t.Fatalf("profile/archetype = %q/%q", r.Profile, r.Archetype)
	}
	if r.DriverModel != "claude-opus" || !r.Headless || r.GeneratedAt != fixedNow {
		t.Fatalf("driver/headless/now = %q/%v/%q", r.DriverModel, r.Headless, r.GeneratedAt)
	}
}

// A nil workflow yields an empty step and no driver model but still assembles.
func TestAssembleVerdict_NilWorkflow(t *testing.T) {
	t.Setenv("CENTINELA_HEADLESS", "")
	deps := fakeDeps([]gates.Result{passGate()}, vr(), nil)
	pkt := AssembleVerdict("feat", &config.Config{}, nil, deps)
	if pkt.Run.Step != "" || pkt.Run.DriverModel != "" {
		t.Fatalf("nil wf must leave step/driver empty, got %q/%q", pkt.Run.Step, pkt.Run.DriverModel)
	}
}
