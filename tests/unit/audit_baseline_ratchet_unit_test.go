package unit_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/audit"
	"github.com/samuelnp/centinela/internal/config"
)

// oversized returns a >100-line Go body so the file_size gate fails.
func oversized(extra int) string {
	var b strings.Builder
	b.WriteString("package big\n")
	for i := 0; i < 110+extra; i++ {
		b.WriteString("// filler line to push the file over the 100-line limit\n")
	}
	return b.String()
}

// auditRepo chdirs into a temp repo with file_size + audit_baseline enabled and
// returns the loaded cfg. Chdir is reverted on cleanup.
func auditRepo(t *testing.T) *config.Config {
	t.Helper()
	dir := t.TempDir()
	if r, err := filepath.EvalSymlinks(dir); err == nil {
		dir = r
	}
	toml := "[gates]\nfile_size = true\n\n[gates.audit_baseline]\nenabled = true\nseverity = \"fail\"\n"
	mustWrite(t, dir, "centinela.toml", toml)
	mustWrite(t, dir, "internal/first.go", oversized(0))
	wd, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config load: %v", err)
	}
	return cfg
}

func mustWrite(t *testing.T, dir, name, body string) {
	t.Helper()
	full := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestRecordThenRatchetLifecycle exercises record → no-new → add-new → remove →
// resolved over a real temp repo.
func TestRecordThenRatchetLifecycle(t *testing.T) {
	cfg := auditRepo(t)

	// Record the current state: first.go is now baselined.
	base := audit.Record(cfg)
	if d := audit.Ratchet(cfg, base); d.HasNew() {
		t.Fatalf("freshly recorded repo should have 0 new, got %d", len(d.New))
	}

	// Introduce a second oversized file → exactly one new, named.
	mustWrite(t, dirOf(t), "internal/second.go", oversized(3))
	d := audit.Ratchet(cfg, base)
	if len(d.New) != 1 {
		t.Fatalf("want 1 new, got %d: %+v", len(d.New), d.New)
	}
	if !strings.Contains(d.New[0].Raw, "internal/second.go") {
		t.Fatalf("new violation not named second.go: %q", d.New[0].Raw)
	}
	if len(d.Baselined) != 1 {
		t.Fatalf("first.go should stay baselined, got %d", len(d.Baselined))
	}

	// Remove the new file → back to 0 new.
	if err := os.Remove(filepath.Join(dirOf(t), "internal/second.go")); err != nil {
		t.Fatal(err)
	}
	if d := audit.Ratchet(cfg, base); d.HasNew() {
		t.Fatalf("after removal expected 0 new, got %d", len(d.New))
	}
}

// TestResolvedReportedWhenBaselinedFileRemoved: deleting a baselined file lists
// it as resolved without producing a new violation.
func TestResolvedReportedWhenBaselinedFileRemoved(t *testing.T) {
	cfg := auditRepo(t)
	base := audit.Record(cfg)
	if err := os.Remove(filepath.Join(dirOf(t), "internal/first.go")); err != nil {
		t.Fatal(err)
	}
	d := audit.Ratchet(cfg, base)
	if d.HasNew() {
		t.Fatalf("fixing a baselined violation must not add new: %+v", d.New)
	}
	if len(d.Resolved) != 1 || !strings.Contains(d.Resolved[0].Raw, "internal/first.go") {
		t.Fatalf("first.go not reported resolved: %+v", d.Resolved)
	}
}

// dirOf returns the current working directory (the active temp repo).
func dirOf(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	return wd
}
