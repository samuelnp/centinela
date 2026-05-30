package memory

import (
	"os"
	"testing"
)

// SC-01: Capture for tests step writes a lesson entry.
func TestCaptureTestsStepWritesLesson(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(dir)        //nolint:errcheck

	os.MkdirAll(".workflow", 0o755) //nolint:errcheck
	os.WriteFile(".workflow/alpha-edge-cases.md", []byte("- timeout edge case\n- retry on 429\n"), 0o644) //nolint:errcheck

	Capture("alpha", "tests", enabledCfg())

	entries := loadEntries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 lesson entry, got %d", len(entries))
	}
	if entries[0].Type != TypeLesson {
		t.Fatalf("expected lesson type, got %q", entries[0].Type)
	}
}

// SC-02: Capture for validate step writes a verdict entry.
func TestCaptureValidateStepWritesVerdict(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(dir)        //nolint:errcheck

	os.MkdirAll(".workflow", 0o755) //nolint:errcheck
	os.WriteFile(".workflow/alpha-gatekeeper.md", []byte("## Gatekeeper\nAll clear. Status: SAFE\n"), 0o644) //nolint:errcheck

	Capture("alpha", "validate", enabledCfg())

	entries := loadEntries()
	if len(entries) != 1 || entries[0].Type != TypeVerdict {
		t.Fatalf("expected 1 verdict entry, got %v", entries)
	}
}

// SC-06: missing artifact does not block — no entry written.
func TestCaptureMissingArtifactNoEntry(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(dir)        //nolint:errcheck

	Capture("alpha", "tests", enabledCfg())

	entries := loadEntries()
	if len(entries) != 0 {
		t.Fatalf("expected no entries when artifact missing, got %d", len(entries))
	}
}

// SC-12: disabled config → no-op.
func TestCaptureDisabledConfig(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(dir)        //nolint:errcheck

	os.MkdirAll(".workflow", 0o755)                                                                       //nolint:errcheck
	os.WriteFile(".workflow/alpha-edge-cases.md", []byte("- lesson\n"), 0o644) //nolint:errcheck

	Capture("alpha", "tests", disabledCfg())

	entries := loadEntries()
	if len(entries) != 0 {
		t.Fatalf("expected no entries when disabled, got %d", len(entries))
	}
}
