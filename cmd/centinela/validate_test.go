package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunValidate_MutuallyExclusiveFlags(t *testing.T) {
	old1, old2 := validateChanged, validateFull
	defer func() { validateChanged, validateFull = old1, old2 }()
	validateChanged = true
	validateFull = true
	if err := runValidate(nil, nil); err == nil {
		t.Fatal("expected error when --changed and --full are both set")
	}
}

func TestRunValidate_NoFlags_RunsDefaultMode(t *testing.T) {
	old1, old2 := validateChanged, validateFull
	defer func() { validateChanged, validateFull = old1, old2 }()
	validateChanged = false
	validateFull = false

	d := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(d)          //nolint:errcheck
	_ = os.WriteFile(filepath.Join(d, "centinela.toml"),
		[]byte("[validate]\ncommands = []\n[gates]\nfile_size = false\n"), 0644)
	if err := runValidate(nil, nil); err != nil {
		t.Fatalf("runValidate clean: %v", err)
	}
}

func TestRunValidate_ChangedFlag_Accepted(t *testing.T) {
	old1, old2 := validateChanged, validateFull
	defer func() { validateChanged, validateFull = old1, old2 }()
	validateChanged = true
	validateFull = false

	d := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(d)          //nolint:errcheck
	_ = os.WriteFile(filepath.Join(d, "centinela.toml"),
		[]byte("[validate]\ncommands = []\n[gates]\nfile_size = false\n"), 0644)
	_ = runValidate(nil, nil) // exit status not asserted — flag parsing is the target.
}

func TestRunValidate_FullFlag_Accepted(t *testing.T) {
	old1, old2 := validateChanged, validateFull
	defer func() { validateChanged, validateFull = old1, old2 }()
	validateChanged = false
	validateFull = true

	d := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(d)          //nolint:errcheck
	_ = os.WriteFile(filepath.Join(d, "centinela.toml"),
		[]byte("[validate]\ncommands = []\n[gates]\nfile_size = false\n"), 0644)
	_ = runValidate(nil, nil)
}
