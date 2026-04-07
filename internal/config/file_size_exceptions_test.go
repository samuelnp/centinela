package config

import (
	"os"
	"strings"
	"testing"
)

func TestLoad_FileSizeExceptionAboveCapFails(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(dir)        //nolint:errcheck

	os.WriteFile(Filename, []byte("[gates]\nfile_size = true\n[[gates.file_size_exceptions]]\npath=\"internal/x.go\"\nkind=\"configuration\"\nreason=\"allowed\"\nmax_lines=140\n"), 0644) //nolint:errcheck
	_, err := Load()
	if err == nil || !strings.Contains(err.Error(), "max_lines") {
		t.Fatalf("expected max_lines validation error, got %v", err)
	}
}

func TestLoad_FileSizeExceptionNormalizesPath(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(dir)        //nolint:errcheck

	os.WriteFile(Filename, []byte("[gates]\nfile_size = true\n[[gates.file_size_exceptions]]\npath=\"internal\\\\x.go\"\nkind=\"domain_atomic\"\nreason=\"single cohesive unit\"\nmax_lines=120\n"), 0644) //nolint:errcheck
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if got := cfg.Gates.FileSizeExceptions[0].Path; got != "internal/x.go" {
		t.Fatalf("expected normalized path, got %q", got)
	}
}
