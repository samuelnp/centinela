package memory

import (
	"os"
	"testing"
)

// SC-12: nil config → no-op.
func TestCaptureNilConfig(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(dir)        //nolint:errcheck

	Capture("alpha", "tests", nil)
	if len(loadEntries()) != 0 {
		t.Fatal("expected no entries for nil config")
	}
}

// Non-capture step (code) is a no-op.
func TestCaptureCodeStepNoOp(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(dir)        //nolint:errcheck

	Capture("alpha", "code", enabledCfg())
	if len(loadEntries()) != 0 {
		t.Fatal("expected no entries for code step")
	}
}
