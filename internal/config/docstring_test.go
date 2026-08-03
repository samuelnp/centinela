package config

import "testing"

func TestNormalizeDocstring_FillsFailSeverityAndInternalDefault(t *testing.T) {
	got := NormalizeDocstring(DocstringConfig{Enabled: true})
	if got.Severity != "fail" {
		t.Fatalf("severity = %q, want fail — a ratcheted gate enforces", got.Severity)
	}
	if !got.IncludesInternal() {
		t.Fatal("include_internal must default to true")
	}
	if got.CheckFields || got.RequireNamePrefix {
		t.Fatal("check_fields and require_name_prefix must default off")
	}
}

func TestNormalizeDocstring_TrimsSeverityAndDropsBlankRoots(t *testing.T) {
	got := NormalizeDocstring(DocstringConfig{
		Enabled:  true,
		Severity: "  warn  ",
		Roots:    []string{" internal ", "", "   ", "cmd"},
	})
	if got.Severity != "warn" {
		t.Fatalf("severity = %q", got.Severity)
	}
	if len(got.Roots) != 2 || got.Roots[0] != "internal" || got.Roots[1] != "cmd" {
		t.Fatalf("roots = %v", got.Roots)
	}
}

func TestNormalizeDocstring_KeepsAnExplicitIncludeInternalFalse(t *testing.T) {
	off := false
	got := NormalizeDocstring(DocstringConfig{Enabled: true, IncludeInternal: &off})
	if got.IncludesInternal() {
		t.Fatal("an explicit include_internal=false must survive normalization")
	}
}

func TestValidateDocstring_RejectsUnknownSeverity(t *testing.T) {
	err := validateDocstring(DocstringConfig{Enabled: true, Severity: "nope"})
	if err == nil {
		t.Fatal("want an error for an unknown severity")
	}
	if got := err.Error(); got != `gates.docstring.severity must be fail or warn, got "nope"` {
		t.Fatalf("error = %q", got)
	}
}

func TestValidateDocstring_AcceptsFailAndWarn(t *testing.T) {
	for _, s := range []string{"fail", "warn"} {
		if err := validateDocstring(DocstringConfig{Enabled: true, Severity: s}); err != nil {
			t.Fatalf("severity %q rejected: %v", s, err)
		}
	}
}

func TestValidateDocstring_DisabledIsANoOp(t *testing.T) {
	if err := validateDocstring(DocstringConfig{Enabled: false, Severity: "nope"}); err != nil {
		t.Fatalf("a disabled gate must not validate severity: %v", err)
	}
}

func TestApplyDefaults_NormalizesTheDocstringBlock(t *testing.T) {
	cfg := &Config{}
	cfg.Gates.Docstring.Enabled = true
	applyDefaults(cfg)
	if cfg.Gates.Docstring.Severity != "fail" || !cfg.Gates.Docstring.IncludesInternal() {
		t.Fatalf("defaults not applied: %+v", cfg.Gates.Docstring)
	}
}

func TestValidateConfig_SurfacesADocstringSeverityTypo(t *testing.T) {
	cfg := &Config{}
	cfg.Gates.Docstring.Enabled = true
	cfg.Gates.Docstring.Severity = "nope"
	if err := validateConfig(cfg); err == nil {
		t.Fatal("validateConfig must reject an unknown docstring severity")
	}
}
