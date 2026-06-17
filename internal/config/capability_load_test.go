package config

import (
	"os"
	"strings"
	"testing"
)

// A bad capability config makes config.Load fail, and the error names the
// offending value — proving validateCapabilities is wired into validateConfig.
func TestLoad_BadCapabilityClassRejected(t *testing.T) {
	t.Chdir(t.TempDir())
	toml := "[orchestration.capabilities]\n\"local/m\" = \"genius\"\n"
	os.WriteFile(Filename, []byte(toml), 0644) //nolint:errcheck
	_, err := Load()
	if err == nil || !strings.Contains(err.Error(), "genius") {
		t.Fatalf("expected capability class rejection naming genius, got %v", err)
	}
}

// A normalized class value (trim + lower) loads successfully and resolves.
func TestLoad_CapabilityClassNormalized(t *testing.T) {
	t.Chdir(t.TempDir())
	toml := "[orchestration.capabilities]\n\"local/m\" = \"  Frontier  \"\n"
	os.WriteFile(Filename, []byte(toml), 0644) //nolint:errcheck
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if class, ok := CapabilityClassFor("local/m", cfg); !ok || class != CapabilityFrontier {
		t.Fatalf("local/m resolved to (%q,%v), want (frontier,true)", class, ok)
	}
}

// An explicit [workflow] enforcement_profile is captured into the raw shadow at
// Load (mirroring RawStepConfirmationMode), so the explicit-global tier engages.
func TestLoad_RawEnforcementProfileCaptured(t *testing.T) {
	t.Chdir(t.TempDir())
	os.WriteFile(Filename, []byte("[workflow]\nenforcement_profile=\"guided\"\n"), 0644) //nolint:errcheck
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if cfg.Workflow.RawEnforcementProfile != ProfileGuided {
		t.Fatalf("raw = %q, want guided", cfg.Workflow.RawEnforcementProfile)
	}

	t.Chdir(t.TempDir())
	os.WriteFile(Filename, []byte("[workflow]\n"), 0644) //nolint:errcheck
	cfg2, _ := Load()
	if cfg2.Workflow.RawEnforcementProfile != "" {
		t.Fatalf("unset profile must leave raw empty, got %q", cfg2.Workflow.RawEnforcementProfile)
	}
}
