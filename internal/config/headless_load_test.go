package config

import (
	"os"
	"path/filepath"
	"testing"
)

// A [headless] section round-trips enabled/detect_ci through config.Load.
func TestLoad_HeadlessSection(t *testing.T) {
	dir := t.TempDir()
	toml := "[headless]\nenabled = true\ndetect_ci = true\n"
	if err := os.WriteFile(filepath.Join(dir, Filename), []byte(toml), 0o644); err != nil {
		t.Fatal(err)
	}
	old, _ := os.Getwd()
	defer os.Chdir(old) //nolint:errcheck
	os.Chdir(dir)       //nolint:errcheck
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !cfg.Headless.Enabled || !cfg.Headless.DetectCI {
		t.Fatalf("headless section not decoded: %+v", cfg.Headless)
	}
}

// Zero-config (no centinela.toml) leaves HeadlessConfig at its false default.
func TestLoad_HeadlessDefaultsOff(t *testing.T) {
	dir := t.TempDir()
	old, _ := os.Getwd()
	defer os.Chdir(old) //nolint:errcheck
	os.Chdir(dir)       //nolint:errcheck
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Headless.Enabled || cfg.Headless.DetectCI {
		t.Fatalf("default headless must be all-false, got %+v", cfg.Headless)
	}
}
