package integration_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/audit"
	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/gates"
)

func oversized(extra int) string {
	var b strings.Builder
	b.WriteString("package big\n")
	for i := 0; i < 110+extra; i++ {
		b.WriteString("// filler line to exceed the 100-line file-size gate limit\n")
	}
	return b.String()
}

// auditRepo chdirs into a temp repo with file_size + audit_baseline at the given
// severity and one oversized file, returning the loaded cfg.
func auditRepo(t *testing.T, severity string) *config.Config {
	t.Helper()
	dir := t.TempDir()
	if r, err := filepath.EvalSymlinks(dir); err == nil {
		dir = r
	}
	toml := "[gates]\nfile_size = true\n\n[gates.audit_baseline]\nenabled = true\nseverity = \"" +
		severity + "\"\n"
	mustWrite(t, dir, "centinela.toml", toml)
	mustWrite(t, dir, "internal/big.go", oversized(0))
	wd, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config load: %v", err)
	}
	return cfg
}

func mustWrite(t *testing.T, dir, name, body string) {
	t.Helper()
	full := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestCheckSkipsWithoutBaseline: the audit_baseline gate is non-blocking until a
// baseline is recorded (safe-adoption default).
func TestCheckSkipsWithoutBaseline(t *testing.T) {
	cfg := auditRepo(t, "fail")
	if r := audit.Check(cfg); r.Status != gates.Skip {
		t.Fatalf("status = %v, want Skip", r.Status)
	}
}

// TestCheckSeverityMapping: after a baseline + a new violation, severity governs
// Fail vs Warn.
func TestCheckSeverityMapping(t *testing.T) {
	for _, tc := range []struct {
		severity string
		want     gates.Status
	}{{"fail", gates.Fail}, {"warn", gates.Warn}} {
		cfg := auditRepo(t, tc.severity)
		// Empty baseline ⇒ the live big.go is new.
		if err := audit.Save(cfg.Gates.AuditBaseline.BaselinePath,
			audit.Baseline{Scheme: "v1", Version: 1}); err != nil {
			t.Fatal(err)
		}
		if r := audit.Check(cfg); r.Status != tc.want {
			t.Fatalf("severity %s: status = %v, want %v", tc.severity, r.Status, tc.want)
		}
	}
}

// TestFingerprintStabilityAcrossGrowth: growing a baselined file by lines keeps
// it baselined, never new (AC-5).
func TestFingerprintStabilityAcrossGrowth(t *testing.T) {
	cfg := auditRepo(t, "fail")
	base := audit.Record(cfg)
	if err := audit.Save(cfg.Gates.AuditBaseline.BaselinePath, base); err != nil {
		t.Fatal(err)
	}

	// Grow big.go by many lines; still oversized, same path identity.
	mustWrite(t, dirOf(t), "internal/big.go", oversized(80))
	d := audit.Ratchet(cfg, base)
	if d.HasNew() {
		t.Fatalf("line growth must not be new: %+v", d.New)
	}
	if len(d.Baselined) != 1 {
		t.Fatalf("want 1 baselined after growth, got %d", len(d.Baselined))
	}
	if r := audit.Check(cfg); r.Status != gates.Pass {
		t.Fatalf("Check after growth = %v, want Pass", r.Status)
	}
}

func dirOf(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	return wd
}
