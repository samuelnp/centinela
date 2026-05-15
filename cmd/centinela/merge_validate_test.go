package main

import (
	"os"
	"path/filepath"
	"testing"
)

// runValidateForMerge returns (passed, output). The failure branch fires when
// executeValidation returns an error (e.g. malformed centinela.toml).
func TestRunValidateForMerge_FailureBranchSurfacesOutput(t *testing.T) {
	d := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(d)          //nolint:errcheck

	// Malformed centinela.toml causes config.Load to error → validation fails.
	_ = os.WriteFile(filepath.Join(d, "centinela.toml"), []byte("not valid toml ===="), 0644)
	passed, out := runValidateForMerge(d)
	if passed {
		t.Fatal("expected validation to fail on malformed centinela.toml")
	}
	if out == "" {
		t.Fatal("expected failure output describing the error")
	}
}

func TestRunValidateForMerge_HappyPath(t *testing.T) {
	d := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(d)          //nolint:errcheck
	_ = os.WriteFile(filepath.Join(d, "centinela.toml"),
		[]byte("[validate]\ncommands = []\n[gates]\nfile_size = false\n"), 0644)
	passed, out := runValidateForMerge(d)
	if !passed {
		t.Fatalf("expected pass, got failure: %s", out)
	}
}
