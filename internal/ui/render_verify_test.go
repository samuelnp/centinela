package ui

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/verify"
)

func TestRenderVerification(t *testing.T) {
	r := verify.VerificationResult{Feature: "demo", Checks: []verify.Check{
		{Claim: "tests-pass", Role: "qa-senior", Status: verify.StatusPass, Detail: "exited 0"},
		{Claim: "coverage-moved", Status: verify.StatusFail, Detail: "claimed 92% measured 78%"},
		{Claim: "edge-cases", Status: verify.StatusWarn, Detail: "no match for timeout"},
		{Claim: "stubs", Status: verify.StatusSkip, Detail: "no outputs"},
	}}
	out := RenderVerification(r)
	for _, want := range []string{"demo", "tests-pass (qa-senior)", "PASS", "FAIL", "WARN", "SKIP",
		"claimed 92% measured 78%", "1 passed, 1 failed, 1 warned, 1 skipped"} {
		if !strings.Contains(out, want) {
			t.Errorf("render missing %q\n%s", want, out)
		}
	}
}

func TestRenderVerificationSummaryColors(t *testing.T) {
	// All-green path: summary present, no failures.
	green := RenderVerification(verify.VerificationResult{Feature: "g", Checks: []verify.Check{
		{Claim: "a", Status: verify.StatusPass},
	}})
	if !strings.Contains(green, "1 passed, 0 failed, 0 warned, 0 skipped") {
		t.Errorf("green summary wrong:\n%s", green)
	}
	// Warn-only path exercises the yellow branch (no fails).
	warn := RenderVerification(verify.VerificationResult{Feature: "w", Checks: []verify.Check{
		{Claim: "a", Status: verify.StatusWarn},
	}})
	if !strings.Contains(warn, "0 failed, 1 warned") {
		t.Errorf("warn summary wrong:\n%s", warn)
	}
}
