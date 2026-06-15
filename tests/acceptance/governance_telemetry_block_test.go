package acceptance_test

import (
	"testing"

	"github.com/samuelnp/centinela/internal/telemetry"
)

// Acceptance: specs/governance-telemetry.feature

// Scenario: Out-of-step write appends a block event with full context
func TestGT_BlockOutOfStep(t *testing.T) {
	gtChdir(t)
	telemetry.RecordBlock(gtCfg(true), "feat", "plan", "tests", "/p/tests/x_test.go", "out-of-step", "")
	evs := gtEvents(t)
	if len(evs) != 1 {
		t.Fatalf("want 1 event, got %d", len(evs))
	}
	e := evs[0]
	if e.Type != telemetry.TypeBlock || e.Reason != "out-of-step" ||
		e.Feature != "feat" || e.Step != "plan" || e.FileType != "tests" ||
		e.TargetPath != "/p/tests/x_test.go" {
		t.Fatalf("block event missing context: %+v", e)
	}
}

// Scenario: Write with no active workflow appends a need-init block event
func TestGT_BlockNeedInit(t *testing.T) {
	gtChdir(t)
	telemetry.RecordBlock(gtCfg(true), "", "", "code", "/p/internal/y.go", "need-init", "")
	evs := gtEvents(t)
	if len(evs) != 1 {
		t.Fatalf("want 1 event, got %d", len(evs))
	}
	e := evs[0]
	if e.Reason != "need-init" || e.Feature != "" || e.Step != "" || e.FileType != "code" {
		t.Fatalf("need-init block wrong: %+v", e)
	}
}
