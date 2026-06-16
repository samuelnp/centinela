package audit

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/gates"
)

// TestCheckMissingBaselineSkips: no baseline file ⇒ non-blocking Skip.
func TestCheckMissingBaselineSkips(t *testing.T) {
	cfg := tempRepo(t, "fail", map[string]string{"internal/big.go": oversizedGo(0)})
	r := Check(cfg)
	if r.Status != gates.Skip {
		t.Fatalf("status = %v, want Skip", r.Status)
	}
}

// TestCheckNewViolationFails: a baseline that omits the live violation ⇒ Fail at
// severity fail, Warn at severity warn.
func TestCheckNewViolationFails(t *testing.T) {
	for _, tc := range []struct {
		severity string
		want     gates.Status
	}{{"fail", gates.Fail}, {"warn", gates.Warn}} {
		cfg := tempRepo(t, tc.severity, map[string]string{"internal/big.go": oversizedGo(0)})
		empty := Baseline{Scheme: fingerprintScheme, Version: 1}
		if err := Save(cfg.Gates.AuditBaseline.BaselinePath, empty); err != nil {
			t.Fatal(err)
		}
		r := Check(cfg)
		if r.Status != tc.want {
			t.Fatalf("severity %s: status = %v, want %v", tc.severity, r.Status, tc.want)
		}
		if len(r.Details) == 0 {
			t.Fatalf("severity %s: expected new-violation details", tc.severity)
		}
	}
}

// TestCheckAllBaselinedPasses: a baseline capturing the live violation ⇒ Pass.
func TestCheckAllBaselinedPasses(t *testing.T) {
	cfg := tempRepo(t, "fail", map[string]string{"internal/big.go": oversizedGo(0)})
	if err := Save(cfg.Gates.AuditBaseline.BaselinePath, Record(cfg)); err != nil {
		t.Fatal(err)
	}
	if r := Check(cfg); r.Status != gates.Pass {
		t.Fatalf("status = %v, want Pass", r.Status)
	}
}

// TestCheckStaleSchemeWarns: a baseline under an old scheme ⇒ non-blocking Warn.
func TestCheckStaleSchemeWarns(t *testing.T) {
	cfg := tempRepo(t, "fail", map[string]string{"internal/big.go": oversizedGo(0)})
	stale := Baseline{Scheme: "v0", Version: 1}
	if err := Save(cfg.Gates.AuditBaseline.BaselinePath, stale); err != nil {
		t.Fatal(err)
	}
	if r := Check(cfg); r.Status != gates.Warn {
		t.Fatalf("status = %v, want Warn", r.Status)
	}
}

// TestCheckCorruptBaselineFails: an unparseable baseline file ⇒ Fail (the load
// error is surfaced, not silently ignored).
func TestCheckCorruptBaselineFails(t *testing.T) {
	cfg := tempRepo(t, "fail", map[string]string{"internal/big.go": oversizedGo(0)})
	path := cfg.Gates.AuditBaseline.BaselinePath
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if r := Check(cfg); r.Status != gates.Fail {
		t.Fatalf("status = %v, want Fail", r.Status)
	}
}
