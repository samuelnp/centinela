package config

import "testing"

// TestNormalizeAuditBaselineDefaults fills unset severity and path with the
// safe-adoption defaults (warn + standard path).
func TestNormalizeAuditBaselineDefaults(t *testing.T) {
	got := NormalizeAuditBaseline(AuditBaselineConfig{})
	if got.Severity != "warn" {
		t.Fatalf("severity = %q, want warn", got.Severity)
	}
	if got.BaselinePath != defaultBaselinePath {
		t.Fatalf("path = %q, want %q", got.BaselinePath, defaultBaselinePath)
	}
}

// TestNormalizeAuditBaselinePreservesCustom keeps explicit values and trims
// surrounding whitespace.
func TestNormalizeAuditBaselinePreservesCustom(t *testing.T) {
	in := AuditBaselineConfig{Severity: " fail ", BaselinePath: " custom.json "}
	got := NormalizeAuditBaseline(in)
	if got.Severity != "fail" {
		t.Fatalf("severity = %q, want fail", got.Severity)
	}
	if got.BaselinePath != "custom.json" {
		t.Fatalf("path = %q, want custom.json", got.BaselinePath)
	}
}

// TestValidateAuditBaselineAcceptsValid passes both legal severities.
func TestValidateAuditBaselineAcceptsValid(t *testing.T) {
	for _, sev := range []string{"fail", "warn"} {
		c := AuditBaselineConfig{Enabled: true, Severity: sev}
		if err := validateAuditBaseline(c); err != nil {
			t.Fatalf("severity %q rejected: %v", sev, err)
		}
	}
}

// TestValidateAuditBaselineRejectsUnknown fails an unknown severity so a typo
// cannot silently change strictness.
func TestValidateAuditBaselineRejectsUnknown(t *testing.T) {
	c := AuditBaselineConfig{Enabled: true, Severity: "block"}
	if err := validateAuditBaseline(c); err == nil {
		t.Fatal("expected error for unknown severity")
	}
}

// TestValidateAuditBaselineDisabledSkips is a no-op when the gate is off, even
// with an otherwise-invalid severity.
func TestValidateAuditBaselineDisabledSkips(t *testing.T) {
	c := AuditBaselineConfig{Enabled: false, Severity: "nonsense"}
	if err := validateAuditBaseline(c); err != nil {
		t.Fatalf("disabled gate should skip validation, got %v", err)
	}
}

// TestApplyDefaultsNormalizesAuditBaseline confirms the section is wired into
// the central applyDefaults pass.
func TestApplyDefaultsNormalizesAuditBaseline(t *testing.T) {
	cfg := &Config{}
	applyDefaults(cfg)
	if cfg.Gates.AuditBaseline.Severity != "warn" {
		t.Fatalf("severity = %q, want warn", cfg.Gates.AuditBaseline.Severity)
	}
	if cfg.Gates.AuditBaseline.BaselinePath != defaultBaselinePath {
		t.Fatalf("path = %q, want %q", cfg.Gates.AuditBaseline.BaselinePath, defaultBaselinePath)
	}
}
