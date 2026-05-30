package verify

import (
	"errors"
	"testing"

	"github.com/samuelnp/centinela/internal/evidence"
)

func TestVerifyNoEvidenceSkips(t *testing.T) {
	// Loader returns nil evidence => no claims; expect a single SKIP placeholder.
	deps := Deps{Runner: &fakeRunner{}, Load: fakeLoad(nil, errors.New("not found"))}
	res := Verify("fresh", "tests", cfgWithCmds("go test"), deps)
	if len(res.Checks) != 1 || res.Checks[0].Status != StatusSkip {
		t.Fatalf("expected single skip, got %+v", res.Checks)
	}
	if res.Checks[0].Detail != "no claims to verify" {
		t.Fatalf("skip detail = %q", res.Checks[0].Detail)
	}
	if res.HasFailures() {
		t.Fatal("no-evidence must not block")
	}
}

func TestVerifyDispatchesFourChecks(t *testing.T) {
	ev := &evidence.RoleEvidence{Coverage: cov(85.0), Outputs: nil, EdgeCases: nil}
	deps := Deps{
		Runner: &fakeRunner{def: RunOutcome{Output: covOut}},
		Load:   fakeLoad(ev, nil),
	}
	res := Verify("f", "tests", cfgWithCmds("go test"), deps)
	if len(res.Checks) != 4 {
		t.Fatalf("expected 4 checks for one role, got %d", len(res.Checks))
	}
	claims := map[string]bool{}
	for _, c := range res.Checks {
		claims[c.Claim] = true
	}
	for _, want := range []string{claimTestsPass, claimCoverage, claimStubs, claimEdgeCases} {
		if !claims[want] {
			t.Errorf("missing check %q", want)
		}
	}
}

func TestVerifyDefaultLoader(t *testing.T) {
	// Load nil => falls back to evidence.Read; in a temp dir with no .workflow
	// the read errors and the role is skipped, yielding the SKIP placeholder.
	deps := Deps{Runner: &fakeRunner{}}
	res := Verify("absent-feature", "tests", cfgWithCmds("go test"), deps)
	if len(res.Checks) != 1 || res.Checks[0].Status != StatusSkip {
		t.Fatalf("default loader on missing evidence => skip, got %+v", res.Checks)
	}
}

func TestVerifyUnknownStepHasNoRoles(t *testing.T) {
	deps := Deps{Runner: &fakeRunner{}, Load: fakeLoad(&evidence.RoleEvidence{}, nil)}
	res := Verify("f", "nonexistent-step", cfgWithCmds("go test"), deps)
	if len(res.Checks) != 1 || res.Checks[0].Status != StatusSkip {
		t.Fatalf("unknown step => skip, got %+v", res.Checks)
	}
}
