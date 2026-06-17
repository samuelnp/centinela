package config

import (
	"strings"
	"testing"
)

// TestNormalizeCustomGatesDefaults fills severity, output, and timeout defaults
// and trims whitespace while leaving explicit values untouched.
func TestNormalizeCustomGatesDefaults(t *testing.T) {
	gs := NormalizeCustomGates([]CustomGate{
		{Name: "  a  ", Command: "  true  "},
		{Name: "b", Command: "false", Severity: "warn", Output: "lines", TimeoutSeconds: 5},
	})
	if gs[0].Name != "a" || gs[0].Command != "true" {
		t.Fatalf("trim failed: %+v", gs[0])
	}
	if gs[0].Severity != "fail" || gs[0].Output != "blob" || gs[0].TimeoutSeconds != 60 {
		t.Fatalf("defaults not applied: %+v", gs[0])
	}
	if gs[1].Severity != "warn" || gs[1].Output != "lines" || gs[1].TimeoutSeconds != 5 {
		t.Fatalf("explicit values mutated: %+v", gs[1])
	}
}

// TestValidateCustomGatesValid accepts a well-formed normalized set.
func TestValidateCustomGatesValid(t *testing.T) {
	gs := NormalizeCustomGates([]CustomGate{
		{Name: "a", Command: "true"},
		{Name: "b", Command: "false", Severity: "warn", Output: "lines"},
	})
	if err := validateCustomGates(gs); err != nil {
		t.Fatalf("valid set rejected: %v", err)
	}
}

// TestValidateCustomGatesErrors maps each malformed entry to its indexed error.
func TestValidateCustomGatesErrors(t *testing.T) {
	cases := []struct {
		name string
		gs   []CustomGate
		want string
	}{
		{"empty name", []CustomGate{{Command: "true", Severity: "fail", Output: "blob"}},
			"gates.custom[0].name is required"},
		{"empty command", []CustomGate{{Name: "x", Severity: "fail", Output: "blob"}},
			"gates.custom[0].command is required"},
		{"duplicate", []CustomGate{
			{Name: "d", Command: "true", Severity: "fail", Output: "blob"},
			{Name: "d", Command: "true", Severity: "fail", Output: "blob"}},
			"gates.custom[1].name \"d\" duplicates gates.custom[0]"},
		{"collision", []CustomGate{{Name: "import_graph", Command: "true", Severity: "fail", Output: "blob"}},
			"collides with built-in gate"},
		{"bad severity", []CustomGate{{Name: "x", Command: "true", Severity: "critical", Output: "blob"}},
			"severity must be fail or warn"},
		{"bad output", []CustomGate{{Name: "x", Command: "true", Severity: "fail", Output: "json"}},
			"output must be blob or lines"},
	}
	for _, c := range cases {
		err := validateCustomGates(c.gs)
		if err == nil || !strings.Contains(err.Error(), c.want) {
			t.Fatalf("%s: err = %v, want contains %q", c.name, err, c.want)
		}
	}
}

// TestValidateCustomGatesValidatesDisabled: a malformed entry is rejected even
// when Enabled is false (typos surface at load time).
func TestValidateCustomGatesValidatesDisabled(t *testing.T) {
	gs := NormalizeCustomGates([]CustomGate{{Enabled: false, Command: "true"}})
	if err := validateCustomGates(gs); err == nil {
		t.Fatal("disabled malformed entry should still error")
	}
}

// TestApplyDefaultsNormalizesCustomGates: applyDefaults wires NormalizeCustomGates
// so a raw config has its custom gates defaulted/trimmed.
func TestApplyDefaultsNormalizesCustomGates(t *testing.T) {
	cfg := &Config{}
	cfg.Gates.CustomGates = []CustomGate{{Name: " g ", Command: " true "}}
	applyDefaults(cfg)
	g := cfg.Gates.CustomGates[0]
	if g.Name != "g" || g.Command != "true" || g.Severity != "fail" ||
		g.Output != "blob" || g.TimeoutSeconds != 60 {
		t.Fatalf("custom gate not normalized through applyDefaults: %+v", g)
	}
}

// TestValidateConfigRejectsBadCustomGate: validateConfig surfaces a malformed
// custom gate's indexed error (the file_size_exceptions.go wiring).
func TestValidateConfigRejectsBadCustomGate(t *testing.T) {
	cfg := defaultConfig()
	cfg.Gates.CustomGates = []CustomGate{{Name: "x", Command: "", Severity: "fail", Output: "blob"}}
	if err := validateConfig(cfg); err == nil ||
		!strings.Contains(err.Error(), "gates.custom[0].command is required") {
		t.Fatalf("validateConfig should reject the bad custom gate, got %v", err)
	}
}
