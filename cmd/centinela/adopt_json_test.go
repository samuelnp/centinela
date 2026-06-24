package main

import (
	"strings"
	"testing"
)

// TestRunAdoptJSONAdopted emits {adopted:true, skipped:false} with the path and
// no human report prose (exit 0).
func TestRunAdoptJSONAdopted(t *testing.T) {
	auditRepo(t)
	out, err := runAdoptCmd(t, false, true)
	if err != nil {
		t.Fatalf("json adopt errored: %v", err)
	}
	for _, want := range []string{"\"adopted\": true", "\"skipped\": false", "audit-baseline.json", "\"per_gate\""} {
		if !strings.Contains(out, want) {
			t.Fatalf("json verdict missing %q: %s", want, out)
		}
	}
	if strings.Contains(out, "Adopted baseline —") {
		t.Fatalf("json mode should not print the human report: %s", out)
	}
}

// TestRunAdoptJSONSkip emits {adopted:false, skipped:true} and exits non-zero
// while leaving the file byte-unchanged.
func TestRunAdoptJSONSkip(t *testing.T) {
	auditRepo(t)
	if _, err := runAdoptCmd(t, false, false); err != nil {
		t.Fatal(err)
	}
	before := adoptBaselineBytes(t)
	out, err := runAdoptCmd(t, false, true)
	if err == nil {
		t.Fatal("json skip should exit non-zero")
	}
	for _, want := range []string{"\"adopted\": false", "\"skipped\": true", "\"per_gate\": {}"} {
		if !strings.Contains(out, want) {
			t.Fatalf("json skip verdict missing %q: %s", want, out)
		}
	}
	if string(before) != string(adoptBaselineBytes(t)) {
		t.Fatal("json skip changed the baseline file")
	}
}
