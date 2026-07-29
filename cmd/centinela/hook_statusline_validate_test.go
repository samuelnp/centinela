package main

import "testing"

func TestClassifyValidateErrorMapsVerifierFailures(t *testing.T) {
	cases := map[string][2]string{
		"gatekeeper verdict: CRITICAL — tokens leak":                        {"VERDICT_CRITICAL", "fix-and-reverify"},
		"gatekeeper verdict is missing or unparseable — expected":           {"VERDICT_UNPARSEABLE", "rerun-verifier"},
		"gatekeeper report has no commands-run record — a narrated verdict": {"MISSING_COMMANDS_RECORD", "rerun-verifier"},
		"gatekeeper verification is stale (verified abc, HEAD is now def)":  {"STALE_VERIFICATION", "reverify-fresh-context"},
		"production readiness: BLOCKING":                                    {"PROD_BLOCKING", "harden-feature"},
		"production readiness report not found":                             {"MISSING_PROD_READINESS", "run-production-readiness"},
	}
	for msg, want := range cases {
		block, next := classifyValidateError(msg)
		if block != want[0] || next != want[1] {
			t.Fatalf("classifyValidateError(%q) = %q/%q, want %q/%q", msg, block, next, want[0], want[1])
		}
	}
}

// A CRITICAL verdict must never be misread as a production-readiness problem,
// which is what the pre-existing substring fallback would have done.
func TestClassifyValidateErrorPrefersVerdictOverBlockingSubstring(t *testing.T) {
	block, _ := classifyValidateError("gatekeeper verdict: CRITICAL — the BLOCKING gate is unrelated")
	if block != "VERDICT_CRITICAL" {
		t.Fatalf("block = %q, want VERDICT_CRITICAL", block)
	}
}
