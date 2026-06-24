package unit_test

import (
	"os"
	"testing"

	"github.com/samuelnp/centinela/internal/audit"
)

// TestAdoptRecordsThenSkips: the first adopt records the current violations and
// the second is skipped (one-time adoption), all over a real temp repo. Reuses
// the audit_baseline_ratchet unit helpers (auditRepo seeds internal/first.go).
func TestAdoptRecordsThenSkips(t *testing.T) {
	cfg := auditRepo(t)

	o, err := audit.Adopt(cfg, false)
	if err != nil {
		t.Fatalf("first adopt: %v", err)
	}
	if o.Skipped || o.Baseline.Total() == 0 {
		t.Fatalf("first adopt should record findings: skipped=%v total=%d", o.Skipped, o.Baseline.Total())
	}
	if _, err := os.Stat(o.Path); err != nil {
		t.Fatalf("baseline not written: %v", err)
	}

	again, err := audit.Adopt(cfg, false)
	if err != nil {
		t.Fatalf("second adopt should not error: %v", err)
	}
	if !again.Skipped {
		t.Fatal("second adopt should be skipped (baseline already exists)")
	}
}
