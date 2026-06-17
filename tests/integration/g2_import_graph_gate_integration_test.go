package integration_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/gates"
)

// Acceptance spec: specs/g2-import-graph-gate.feature
//
// Integration coverage runs the gate end-to-end (real `go list -json`) against a
// throwaway module on disk via the public gates.RunWithFilter surface.

func writeModule(t *testing.T, aBody string) {
	t.Helper()
	d := t.TempDir()
	orig, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(orig) }) //nolint:errcheck
	write := func(rel, body string) {
		p := filepath.Join(d, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("go.mod", "module igfix\n\ngo 1.21\n")
	write("b/b.go", "package b\n\nfunc B() {}\n")
	write("a/a.go", aBody)
	if err := os.Chdir(d); err != nil {
		t.Fatal(err)
	}
}

func igGateCfg() *config.Config { //nolint:dupl
	cfg := &config.Config{}
	cfg.Gates.ImportGraph = config.ImportGraphConfig{
		Enabled: true, Module: "igfix",
		Layers: []config.Layer{
			{Name: "leaf", Paths: []string{"b/**"}},
			{Name: "top", Paths: []string{"a/**"}}, // top may NOT import leaf
		},
	}
	return cfg
}

func graphResult(t *testing.T) gates.Result {
	t.Helper()
	for _, r := range gates.RunWithFilter(igGateCfg(), nil) {
		if r.Name == "import_graph" {
			return r
		}
	}
	t.Fatal("import_graph result missing")
	return gates.Result{}
}

// Scenario: All imports respect the matrix -> Pass.
func TestImportGraph_Integration_CleanPasses(t *testing.T) {
	writeModule(t, "package a\n")
	if r := graphResult(t); r.Status != gates.Pass {
		t.Fatalf("clean module should Pass, got %v: %q", r.Status, r.Message)
	}
}

// Scenario Outline: forbidden cross-layer edge -> Fail with the arrow message.
func TestImportGraph_Integration_ForbiddenEdgeFails(t *testing.T) {
	writeModule(t, "package a\n\nimport _ \"igfix/b\"\n")
	r := graphResult(t)
	if r.Status != gates.Fail {
		t.Fatalf("forbidden edge should Fail, got %v", r.Status)
	}
	if !strings.Contains(strings.Join(r.Details, "\n"), "a -> b (top may not import leaf)") {
		t.Fatalf("violating edge not reported: %v", r.Details)
	}
}
