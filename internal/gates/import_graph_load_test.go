package gates

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

// writeFixtureModule creates a minimal two-package module under t.TempDir and
// chdirs into it. pkgBody is the source of package "a" (which may import "b").
func writeFixtureModule(t *testing.T, pkgBody string) {
	t.Helper()
	d := t.TempDir()
	orig, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(orig) }) //nolint:errcheck
	mk := func(rel, body string) {
		p := filepath.Join(d, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	mk("go.mod", "module fixturemod\n\ngo 1.21\n")
	mk("b/b.go", "package b\n\nfunc B() {}\n")
	mk("a/a.go", pkgBody)
	if err := os.Chdir(d); err != nil {
		t.Fatal(err)
	}
}

// writeBrokenModule chdirs into a temp dir whose go.mod is malformed so any
// `go list` invocation exits non-zero.
func writeBrokenModule(t *testing.T) {
	t.Helper()
	d := t.TempDir()
	orig, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(orig) }) //nolint:errcheck
	if err := os.WriteFile(filepath.Join(d, "go.mod"), []byte("not a valid go.mod\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(d); err != nil {
		t.Fatal(err)
	}
}

func fixtureLayers() []config.Layer {
	return []config.Layer{
		{Name: "leaf", Paths: []string{"b/**"}, Allow: nil},
		{Name: "top", Paths: []string{"a/**"}, Allow: nil}, // a may NOT import b
	}
}

func TestRunImportGraph_FailOnForbiddenEdge(t *testing.T) {
	writeFixtureModule(t, "package a\n\nimport _ \"fixturemod/b\"\n")
	cfg := igCfg("fixturemod", fixtureLayers())
	r := checkImportGraph(cfg, nil)
	if r.Status != Fail {
		t.Fatalf("expected Fail on forbidden edge, got %v: %q", r.Status, r.Message)
	}
	joined := strings.Join(r.Details, "\n")
	if !strings.Contains(joined, "a -> b (top may not import leaf)") {
		t.Fatalf("forbidden edge not reported: %q", joined)
	}
}

func TestRunImportGraph_LoadErrorFails(t *testing.T) {
	// A malformed go.mod makes `go list -json` exit non-zero -> Fail with the
	// load diagnostic folded in (never a false Pass on unloadable code).
	writeBrokenModule(t)
	r := checkImportGraph(igCfg("fixturemod", fixtureLayers()), nil)
	if r.Status != Fail {
		t.Fatalf("expected Fail on load error, got %v: %q", r.Status, r.Message)
	}
	if !strings.HasPrefix(r.Message, "import_graph: ") {
		t.Fatalf("load error should be prefixed import_graph:, got %q", r.Message)
	}
}

func TestImportGraph_DiscoveryErrorFails(t *testing.T) {
	// Blank module + broken go.mod -> the go provider's `go list -m` discovery
	// fails -> provider load error -> Fail prefixed "import_graph: ".
	writeBrokenModule(t)
	r := checkImportGraph(igCfg("", fixtureLayers()), nil)
	if r.Status != Fail || !strings.HasPrefix(r.Message, "import_graph: ") {
		t.Fatalf("expected load-error Fail, got %v: %q", r.Status, r.Message)
	}
}
