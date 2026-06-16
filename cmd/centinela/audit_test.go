package main

import (
	"strings"
	"testing"
)

// TestRunAuditMissingBaseline prints the safe-adoption hint and returns nil.
func TestRunAuditMissingBaseline(t *testing.T) {
	auditRepo(t)
	out, err := runCmd(t, false, runAudit)
	if err != nil {
		t.Fatalf("missing baseline should not error: %v", err)
	}
	if !strings.Contains(out, "no baseline") {
		t.Fatalf("expected hint, got %q", out)
	}
}

// TestRunAuditBaselineThenClean records a baseline then audits clean (0 new).
func TestRunAuditBaselineThenClean(t *testing.T) {
	auditRepo(t)
	bOut, err := runCmd(t, false, runAuditBaseline)
	if err != nil || !strings.Contains(bOut, "baselined") {
		t.Fatalf("baseline: out=%q err=%v", bOut, err)
	}
	out, err := runCmd(t, false, runAudit)
	if err != nil {
		t.Fatalf("clean audit errored: %v", err)
	}
	if !strings.Contains(out, "0 new") {
		t.Fatalf("expected 0 new, got %q", out)
	}
	// JSON clean path: verdict with 0 new and no error.
	jOut, jErr := runCmd(t, true, runAudit)
	if jErr != nil {
		t.Fatalf("clean json audit errored: %v", jErr)
	}
	if !strings.Contains(jOut, "\"new\": 0") {
		t.Fatalf("expected json 0 new, got %q", jOut)
	}
}

// TestRunAuditNewViolationBlocks returns a non-nil error when a new violation
// appears, and --json emits a verdict.
func TestRunAuditNewViolationBlocks(t *testing.T) {
	dir := auditRepo(t)
	if _, err := runCmd(t, false, runAuditBaseline); err != nil {
		t.Fatal(err)
	}
	writeAudit(t, dir, "internal/new.go", auditOversizedBody())
	if _, err := runCmd(t, false, runAudit); err == nil {
		t.Fatal("new violation should error (non-zero exit)")
	}
	out, err := runCmd(t, true, runAudit)
	if err == nil {
		t.Fatal("json path should also error on new")
	}
	if !strings.Contains(out, "\"new\": 1") {
		t.Fatalf("json verdict missing new count: %q", out)
	}
}

// TestRunAuditConfigError surfaces a config-load failure from both commands.
func TestRunAuditConfigError(t *testing.T) {
	dir := auditRepo(t)
	writeAudit(t, dir, "centinela.toml", "this = = bad toml")
	if _, err := runCmd(t, false, runAudit); err == nil {
		t.Fatal("runAudit should propagate config error")
	}
	if _, err := runCmd(t, false, runAuditBaseline); err == nil {
		t.Fatal("runAuditBaseline should propagate config error")
	}
}

// TestRunAuditJSONNoBaseline emits an empty JSON verdict and no error.
func TestRunAuditJSONNoBaseline(t *testing.T) {
	auditRepo(t)
	out, err := runCmd(t, true, runAudit)
	if err != nil {
		t.Fatalf("no-baseline json should not error: %v", err)
	}
	if !strings.Contains(out, "\"new\": 0") {
		t.Fatalf("expected empty verdict: %q", out)
	}
}
