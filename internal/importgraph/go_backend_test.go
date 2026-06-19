package importgraph

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGoProvider_LoadExplicitModule(t *testing.T) {
	// Runs within this repo (the importgraph package dir): `go list ./...`
	// resolves real packages, scoped against the explicit module.
	g, err := goProvider{module: "github.com/samuelnp/centinela"}.Load("")
	if err != nil {
		t.Fatal(err)
	}
	if g.Module != "github.com/samuelnp/centinela" || len(g.Pkgs) == 0 {
		t.Fatalf("expected scoped pkgs, got module=%q pkgs=%d", g.Module, len(g.Pkgs))
	}
}

func TestGoProvider_LoadBlankDiscoversModule(t *testing.T) {
	g, err := goProvider{}.Load("")
	if err != nil || g.Module == "" {
		t.Fatalf("blank module must be discovered via go list -m: %v", err)
	}
}

func TestGoProvider_DiscoveryErrorOnBrokenModule(t *testing.T) {
	chdirBroken(t)
	if _, err := (goProvider{}).Load(""); err == nil {
		t.Fatal("blank module + broken go.mod must error on `go list -m`")
	}
}

func TestGoProvider_PackagesErrorOnBrokenModule(t *testing.T) {
	chdirBroken(t)
	if _, err := (goProvider{module: "m"}).Load(""); err == nil {
		t.Fatal("explicit module + broken go.mod must error on `go list -json`")
	}
}

func TestGoProvider_Name(t *testing.T) {
	if (goProvider{}).Name() != "go" {
		t.Fatal("name")
	}
}

// chdirBroken chdirs into a temp dir with a malformed go.mod so any `go list`
// invocation exits non-zero, restoring the original CWD on cleanup.
func chdirBroken(t *testing.T) {
	t.Helper()
	d := t.TempDir()
	if err := os.WriteFile(filepath.Join(d, "go.mod"), []byte("not valid\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	orig, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(orig) }) //nolint:errcheck
	if err := os.Chdir(d); err != nil {
		t.Fatal(err)
	}
}
