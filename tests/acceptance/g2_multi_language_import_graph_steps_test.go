package acceptance_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/gates"
)

// igLayers builds an import_graph config whose matrix forbids a -> b: layer
// "top" (a/**) may not import layer "leaf" (b/**). module is the Go module
// override (ignored by non-go providers).
func igLayers(module string) config.ImportGraphConfig {
	return config.ImportGraphConfig{
		Enabled: true,
		Module:  module,
		Layers: []config.Layer{
			{Name: "leaf", Paths: []string{"b/**"}},
			{Name: "top", Paths: []string{"a/**"}},
		},
	}
}

// runImportGate runs the enabled gates from dir and returns the import_graph
// Result. Only import_graph is enabled, so it is the sole result.
func runImportGate(t *testing.T, dir string, ig config.ImportGraphConfig) gates.Result {
	t.Helper()
	orig, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(orig) }) //nolint:errcheck
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{}
	cfg.Gates.ImportGraph = ig
	for _, r := range gates.RunAll(cfg) {
		if r.Name == "import_graph" {
			return r
		}
	}
	t.Fatal("no import_graph result")
	return gates.Result{}
}

// writeGoFixture creates a two-package module (module fixturemod) where a
// imports b, and returns its directory.
func writeGoFixture(t *testing.T) string {
	t.Helper()
	d := t.TempDir()
	g2write(t, d, "go.mod", "module fixturemod\n\ngo 1.21\n")
	g2write(t, d, "b/b.go", "package b\n\nfunc B() {}\n")
	g2write(t, d, "a/a.go", "package a\n\nimport _ \"fixturemod/b\"\n")
	return d
}

// writeEdgeScript writes an executable script that emits the import-graph JSON
// contract (a imports b) and returns the argv to run it.
func writeEdgeScript(t *testing.T, dir string) []string {
	t.Helper()
	p := filepath.Join(dir, "edges.sh")
	body := "#!/bin/sh\n" +
		`echo '{"module":"m","pkgs":[{"path":"a","imports":["b"]},{"path":"b","imports":[]}]}'` + "\n"
	g2write(t, dir, "edges.sh", body)
	if err := os.Chmod(p, 0o755); err != nil {
		t.Fatal(err)
	}
	return []string{"sh", p}
}

func g2write(t *testing.T, dir, rel, body string) {
	t.Helper()
	p := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func toolPresent(names ...string) bool {
	for _, n := range names {
		if _, err := exec.LookPath(n); err == nil {
			return true
		}
	}
	return false
}
