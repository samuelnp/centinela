package main

import (
	"strings"
	"testing"
)

// completedValidationOutcome is the synthesized machine record of the
// validate gate's own executeValidation run. It must read as an
// unambiguous PASS (exit 0, no start error, no timeout) and its Output
// must name the verified tree so verification transcripts stay honest.
func TestCompletedValidationOutcome_IsPassingRecordNamingVerifiedTree(t *testing.T) {
	outcome := completedValidationOutcome()
	if outcome == nil {
		t.Fatal("outcome must be non-nil")
	}
	if outcome.ExitCode != 0 {
		t.Fatalf("ExitCode = %d, want 0", outcome.ExitCode)
	}
	if outcome.StartErr != nil {
		t.Fatalf("StartErr = %v, want nil", outcome.StartErr)
	}
	if outcome.TimedOut {
		t.Fatal("TimedOut = true, want false")
	}
	if strings.TrimSpace(outcome.Output) == "" {
		t.Fatal("Output must be non-empty")
	}
	if !strings.Contains(outcome.Output, "verified tree") {
		t.Fatalf("Output %q must name the verified tree", outcome.Output)
	}
}
