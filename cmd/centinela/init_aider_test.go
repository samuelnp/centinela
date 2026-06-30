package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestSetupAider_CreatesManagedFiles covers the registry-driven Aider setup
// path: it writes the managed AGENTS.md + .aider.conf.yml and is idempotent.
func TestSetupAider_CreatesManagedFiles(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	if err := setupAider(); err != nil {
		t.Fatalf("setupAider: %v", err)
	}
	for _, want := range []string{"AGENTS.md", ".aider.conf.yml"} {
		if _, err := os.Stat(want); err != nil {
			t.Errorf("missing %s: %v", want, err)
		}
	}
	// Idempotent re-run must not error (hits the no-op managed-marker branch).
	if err := setupAider(); err != nil {
		t.Fatalf("second setupAider must be a no-op: %v", err)
	}
}

// TestSetupAider_ManualReviewPreservesUnmanaged covers the manual-review branch:
// a pre-existing UNMANAGED .aider.conf.yml is preserved, never clobbered.
func TestSetupAider_ManualReviewPreservesUnmanaged(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o) //nolint:errcheck
	os.Chdir(d)       //nolint:errcheck

	const sentinel = "read: my-own-rules.md\n"
	if err := os.WriteFile(filepath.Join(d, ".aider.conf.yml"), []byte(sentinel), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if err := setupAider(); err != nil {
		t.Fatalf("setupAider with unmanaged file: %v", err)
	}
	got, _ := os.ReadFile(filepath.Join(d, ".aider.conf.yml"))
	if string(got) != sentinel {
		t.Errorf("unmanaged .aider.conf.yml must be preserved, got: %q", string(got))
	}
}
