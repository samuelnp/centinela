package config

import (
	"os"
	"path/filepath"
	"testing"
)

func loadInDir(t *testing.T, toml string) *Config {
	t.Helper()
	dir := t.TempDir()
	orig, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(orig) })
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	if toml != "" {
		if err := os.WriteFile(filepath.Join(dir, Filename), []byte(toml), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	return cfg
}

func TestVerifyConfigDefaults(t *testing.T) {
	cfg := loadInDir(t, "[validate]\ncommands=[\"go test\"]\n")
	if cfg.Verify.TimeoutSeconds != 60 {
		t.Errorf("default verify_timeout = %d, want 60", cfg.Verify.TimeoutSeconds)
	}
	if cfg.Verify.CoverageTolerance != 0.001 {
		t.Errorf("default coverage_tolerance = %v, want 0.001", cfg.Verify.CoverageTolerance)
	}
}

func TestVerifyConfigOverrides(t *testing.T) {
	cfg := loadInDir(t, "[verify]\nverify_timeout=30\ncoverage_tolerance=0.5\n")
	if cfg.Verify.TimeoutSeconds != 30 {
		t.Errorf("verify_timeout = %d, want 30", cfg.Verify.TimeoutSeconds)
	}
	if cfg.Verify.CoverageTolerance != 0.5 {
		t.Errorf("coverage_tolerance = %v, want 0.5", cfg.Verify.CoverageTolerance)
	}
}

func TestVerifyConfigNonPositiveResetsToDefault(t *testing.T) {
	cfg := loadInDir(t, "[verify]\nverify_timeout=-1\ncoverage_tolerance=-2\n")
	if cfg.Verify.TimeoutSeconds != 60 || cfg.Verify.CoverageTolerance != 0.001 {
		t.Errorf("non-positive should reset: %+v", cfg.Verify)
	}
}
