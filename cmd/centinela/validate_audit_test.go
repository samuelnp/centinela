package main

import (
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/gates"
)

// TestAppendAuditGateDisabled is a no-op when the gate is off.
func TestAppendAuditGateDisabled(t *testing.T) {
	cfg := &config.Config{}
	cfg.Gates.AuditBaseline.Enabled = false
	got := appendAuditGate(cfg, nil)
	if len(got) != 0 {
		t.Fatalf("disabled gate should append nothing, got %d", len(got))
	}
}

// TestAppendAuditGateEnabled appends one audit_baseline result when enabled.
// With no baseline on disk the gate yields a non-blocking Skip.
func TestAppendAuditGateEnabled(t *testing.T) {
	auditRepo(t) // chdir into a repo with the gate enabled
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	got := appendAuditGate(cfg, []gates.Result{})
	if len(got) != 1 {
		t.Fatalf("enabled gate should append one result, got %d", len(got))
	}
	if got[0].Status != gates.Skip {
		t.Fatalf("missing baseline should Skip, got %v", got[0].Status)
	}
}
