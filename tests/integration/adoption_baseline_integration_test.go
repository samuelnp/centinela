package integration_test

import (
	"testing"

	"github.com/samuelnp/centinela/internal/audit"
	"github.com/samuelnp/centinela/internal/gates"
)

// TestAdoptThenRatchetClean: after adopting the current violations, a fresh
// ratchet over the unchanged repo reports zero new — day-one validate is not
// drowned by the pre-existing debt. Reuses the audit_baseline_ratchet
// integration helpers (auditRepo seeds an oversized internal/big.go).
func TestAdoptThenRatchetClean(t *testing.T) {
	cfg := auditRepo(t, "fail")

	o, err := audit.Adopt(cfg, false)
	if err != nil {
		t.Fatalf("adopt: %v", err)
	}
	if o.Skipped || o.Baseline.Total() == 0 {
		t.Fatalf("adopt should record findings: skipped=%v total=%d", o.Skipped, o.Baseline.Total())
	}

	base, exists, err := audit.Load(cfg.Gates.AuditBaseline.BaselinePath)
	if err != nil || !exists {
		t.Fatalf("load adopted baseline: exists=%v err=%v", exists, err)
	}
	if d := audit.Ratchet(cfg, base); d.HasNew() {
		t.Fatalf("post-adoption ratchet should report 0 new, got %d", len(d.New))
	}
	if r := audit.Check(cfg); r.Status == gates.Fail {
		t.Fatal("audit_baseline gate should not fail right after adoption")
	}
}
