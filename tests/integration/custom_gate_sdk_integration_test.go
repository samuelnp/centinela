package integration_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/samuelnp/centinela/internal/audit"
	"github.com/samuelnp/centinela/internal/config"
)

// customLinesRepo chdirs into a temp repo with a single failing output=lines
// custom gate whose command prints the supplied violation lines, and returns the
// loaded config plus the repo dir. Reloading after rewriteCustom picks up edits.
func customLinesRepo(t *testing.T, printf string) (*config.Config, string) {
	t.Helper()
	dir := t.TempDir()
	if r, err := filepath.EvalSymlinks(dir); err == nil {
		dir = r
	}
	wd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(wd) })
	rewriteCustom(t, dir, printf)
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config load: %v", err)
	}
	return cfg, dir
}

func rewriteCustom(t *testing.T, dir, printf string) {
	t.Helper()
	toml := "[gates]\nfile_size = false\ni18n = true\n\n" +
		"[gates.audit_baseline]\nenabled = true\nseverity = \"fail\"\n\n" +
		"[[gates.custom]]\nenabled = true\nname = \"per-line\"\n" +
		"command = \"printf '" + printf + "'; exit 1\"\noutput = \"lines\"\n"
	if err := os.WriteFile(filepath.Join(dir, "centinela.toml"), []byte(toml), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestCustomLinesBaselineRatchetsNewLine is the AC-4 cross-feature integration:
// a failing output=lines custom gate is baselined per-line; a later run with an
// extra line tolerates the baselined two and reports only the new line.
func TestCustomLinesBaselineRatchetsNewLine(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX sh assumed")
	}
	cfg, dir := customLinesRepo(t, `a.go:1\nb.go:2\n`)

	base := audit.Record(cfg)
	// Each violation line is fingerprinted individually.
	var lines int
	for _, e := range base.Gates {
		if e.Gate == "per-line" {
			lines = len(e.Fingerprints)
		}
	}
	if lines != 2 {
		t.Fatalf("baseline should hold 2 per-line fingerprints, got %d", lines)
	}

	// A new violation line appears; the baseline is unchanged. Reload so the
	// gate's command reflects the rewritten config.
	rewriteCustom(t, dir, `a.go:1\nb.go:2\nc.go:3\n`)
	cfg2, err := config.Load()
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	d := audit.Ratchet(cfg2, base)
	if len(d.New) != 1 {
		t.Fatalf("want exactly 1 new line, got %d: %+v", len(d.New), d.New)
	}
	if len(d.Baselined) != 2 {
		t.Fatalf("want 2 baselined lines, got %d", len(d.Baselined))
	}
	if d.New[0].Raw != "c.go:3" {
		t.Fatalf("new line should be c.go:3, got %q", d.New[0].Raw)
	}
}
