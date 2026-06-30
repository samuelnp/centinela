package acceptance_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/samuelnp/centinela/internal/selfupdate"
)

// Acceptance: specs/cli-self-update.feature
// Scenario: Update resolves a symlinked binary to its real path before replacing
func TestCliSelfUpdate_SymlinkResolution(t *testing.T) {
	dir := t.TempDir()
	real := filepath.Join(dir, "real", "centinela")
	link := filepath.Join(dir, "link", "centinela")
	if err := os.MkdirAll(filepath.Dir(real), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(link), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(real, []byte("OLD"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(real, link); err != nil {
		t.Fatal(err)
	}
	t.Setenv("XDG_CACHE_HOME", filepath.Join(dir, "cache"))
	srv := newAcServer(t, acFakeOpts{tag: "v0.40.2", withAsset: true, goodSum: true})
	u := &selfupdate.Updater{
		Version: "0.37.0",
		GOOS:    runtime.GOOS,
		GOARCH:  runtime.GOARCH,
		APIBase: srv.URL,
		HTTP:    srv.Client(),
		TTL:     time.Hour,
		Now:     time.Now,
		Target:  func() (string, error) { return filepath.EvalSymlinks(link) },
	}
	if _, err := u.Update(); err != nil {
		t.Fatalf("update: %v", err)
	}
	if got, _ := os.ReadFile(real); string(got) != "NEW" {
		t.Fatalf("real binary not updated: %q", got)
	}
	if got, _ := os.ReadFile(link); string(got) != "NEW" {
		t.Fatalf("symlink content stale: %q", got)
	}
}
