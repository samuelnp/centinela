package config

import (
	"os"
	"testing"
)

func TestLoad_DefaultWhenMissing(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(dir)        //nolint:errcheck

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if !cfg.Gates.FileSizeEnabled || cfg.Gates.I18nEnabled {
		t.Fatalf("unexpected defaults: %+v", cfg.Gates)
	}
}

func TestLoad_ParseAndDefaults(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(dir)        //nolint:errcheck

	os.WriteFile(Filename, []byte("[gates]\ni18n = true\n"), 0644) //nolint:errcheck
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if cfg.Gates.FileSizeEnabled || !cfg.Gates.I18nEnabled {
		t.Fatalf("unexpected gate values: %+v", cfg.Gates)
	}
}

func TestApplyDefaultsKeepsConfiguredFileSize(t *testing.T) {
	cfg := &Config{Gates: GatesConfig{FileSizeEnabled: true, I18nEnabled: false}}
	applyDefaults(cfg)
	if !cfg.Gates.FileSizeEnabled {
		t.Fatal("file size gate should remain enabled")
	}
}

func TestApplyDefaultsEnablesFileSizeWhenBothDisabled(t *testing.T) {
	cfg := &Config{Gates: GatesConfig{FileSizeEnabled: false, I18nEnabled: false}}
	applyDefaults(cfg)
	if !cfg.Gates.FileSizeEnabled {
		t.Fatal("file size gate should default to enabled")
	}
}

func TestLoad_ParseError(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(dir)        //nolint:errcheck

	os.WriteFile(Filename, []byte("[gates\n"), 0644) //nolint:errcheck
	if _, err := Load(); err == nil {
		t.Fatal("expected parse error")
	}
}

func TestLoad_ReadError(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(dir)        //nolint:errcheck

	os.Mkdir(Filename, 0755) //nolint:errcheck
	if _, err := Load(); err == nil {
		t.Fatal("expected read error for directory config path")
	}
}
