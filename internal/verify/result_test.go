package verify

import "testing"

func TestStatusBlocking(t *testing.T) {
	blocking := []Status{StatusFail, StatusConfigError, StatusTimeout}
	for _, s := range blocking {
		if !s.blocking() {
			t.Errorf("status %q should block", s)
		}
	}
	for _, s := range []Status{StatusPass, StatusSkip, StatusWarn} {
		if s.blocking() {
			t.Errorf("status %q should not block", s)
		}
	}
}

func TestResultHelpers(t *testing.T) {
	r := VerificationResult{Feature: "f", Checks: []Check{
		{Claim: "a", Status: StatusPass},
		{Claim: "b", Status: StatusFail},
		{Claim: "c", Status: StatusSkip},
		{Claim: "d", Status: StatusWarn},
	}}
	if !r.HasFailures() {
		t.Error("expected HasFailures true")
	}
	if !r.HasWarnings() {
		t.Error("expected HasWarnings true")
	}
	if got := r.Failed(); len(got) != 1 || got[0].Claim != "b" {
		t.Errorf("Failed() = %v", got)
	}
	pass, fail, skip, warn := r.Tally()
	if pass != 1 || fail != 1 || skip != 1 || warn != 1 {
		t.Errorf("Tally() = %d,%d,%d,%d", pass, fail, skip, warn)
	}
}

func TestResultHelpersClean(t *testing.T) {
	r := VerificationResult{Checks: []Check{{Status: StatusPass}, {Status: StatusSkip}}}
	if r.HasFailures() || r.HasWarnings() {
		t.Error("clean result should have no failures or warnings")
	}
	if got := r.Failed(); got != nil {
		t.Errorf("Failed() = %v, want nil", got)
	}
	// Config-error and timeout both count toward the fail tally bucket.
	r2 := VerificationResult{Checks: []Check{{Status: StatusConfigError}, {Status: StatusTimeout}}}
	if _, fail, _, _ := r2.Tally(); fail != 2 {
		t.Errorf("expected 2 fails, got %d", fail)
	}
}
