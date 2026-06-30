package gates

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNewReportFile_TempDirUnwritableReturnsError forces os.CreateTemp to fail
// by pointing TMPDIR at a non-existent path, covering newReportFile's error arm
// (it must return an empty path and a no-op cleanup, never a usable file).
func TestNewReportFile_TempDirUnwritableReturnsError(t *testing.T) {
	t.Setenv("TMPDIR", filepath.Join(t.TempDir(), "does-not-exist"))
	path, cleanup, err := newReportFile()
	if err == nil {
		t.Fatal("expected error when temp dir is unwritable")
	}
	if path != "" {
		t.Fatalf("expected empty path on error, got %q", path)
	}
	cleanup() // must be a safe no-op
}

// TestCheckSecrets_ReportFileErrorYieldsWarn verifies that, with gitleaks
// present but no writable temp dir, checkSecrets degrades to Warn (never a
// false Pass) instead of crashing.
func TestCheckSecrets_ReportFileErrorYieldsWarn(t *testing.T) {
	dir := makeFakeBin(t, "gitleaks", "exit 0")
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("TMPDIR", filepath.Join(t.TempDir(), "nope"))
	r := checkSecrets(emptyPathCfg(), nil)
	if r.Status != Warn {
		t.Fatalf("report-file error must yield Warn, got %v: %q", r.Status, r.Message)
	}
	if !strings.Contains(r.Message, "report file") {
		t.Fatalf("Warn message must name the report-file failure, got %q", r.Message)
	}
}
