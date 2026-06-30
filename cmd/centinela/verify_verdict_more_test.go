package main

import (
	"os"
	"strings"
	"testing"
)

// TestRunVerifyWarnsReturnsError covers the non-clean exit: an edge case with no
// matching test name produces a WARN, so runVerify returns the sentinel error.
func TestRunVerifyWarnsReturnsError(t *testing.T) {
	dishonest := `{"feature":"feat","step":"tests","role":"qa-senior","status":"done",` +
		`"generatedAt":"2026-05-29T00:00:00Z","inputs":["i"],"outputs":[],` +
		`"edgeCases":["zzqqxx unmatched phantom scenario"],"handoffTo":"validation-specialist"}`
	setupVerifyDir(t, dishonest)
	err := runVerify(nil, []string{"feat"})
	if err == nil || !strings.Contains(err.Error(), "did not pass cleanly") {
		t.Fatalf("unmatched edge case should fail verification, got %v", err)
	}
}

// TestRunVerdictMissingWorkflow covers the workflow.Load error branch: no
// .workflow/<feature>.json present → runVerdict surfaces the load error.
func TestRunVerdictMissingWorkflow(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.WriteFile("centinela.toml", []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := runVerdict(nil, []string{"ghost"}); err == nil {
		t.Fatal("expected workflow load error for missing feature")
	}
}
