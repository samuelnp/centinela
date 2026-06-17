package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/gates"
)

// Acceptance for specs/g2-import-graph-gate.feature. Drives the real gate
// (go list -json) against on-disk fixture modules through the public
// gates.RunWithFilter surface, asserting the spec's result semantics.

func igFixture(t *testing.T, aBody string) {
	t.Helper()
	d := t.TempDir()
	orig, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(orig) }) //nolint:errcheck
	w := func(rel, body string) {
		p := filepath.Join(d, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	w("go.mod", "module acc\n\ngo 1.21\n")
	w("b/b.go", "package b\n\nfunc B() {}\n")
	w("a/a.go", aBody)
	if err := os.Chdir(d); err != nil {
		t.Fatal(err)
	}
}

func igAccConfig(layers []config.Layer) *config.Config {
	cfg := &config.Config{}
	cfg.Gates.ImportGraph = config.ImportGraphConfig{Enabled: true, Module: "acc", Layers: layers}
	return cfg
}

func igAccResult(t *testing.T, cfg *config.Config) gates.Result {
	t.Helper()
	for _, r := range gates.RunWithFilter(cfg, nil) {
		if r.Name == "import_graph" {
			return r
		}
	}
	t.Fatal("import_graph result missing")
	return gates.Result{}
}

func twoLayers() []config.Layer {
	return []config.Layer{
		{Name: "leaf", Paths: []string{"b/**"}},
		{Name: "top", Paths: []string{"a/**"}}, // top may NOT import leaf
	}
}

// Scenario: All imports respect the layer matrix — gate passes.
func TestAccept_ImportGraph_CleanPasses(t *testing.T) {
	igFixture(t, "package a\n")
	if r := igAccResult(t, igAccConfig(twoLayers())); r.Status != gates.Pass {
		t.Fatalf("expected Pass, got %v: %q", r.Status, r.Message)
	}
}

// Scenario Outline: A package imports a layer it may not — gate fails with the edge.
func TestAccept_ImportGraph_ForbiddenEdgeFails(t *testing.T) {
	igFixture(t, "package a\n\nimport _ \"acc/b\"\n")
	r := igAccResult(t, igAccConfig(twoLayers()))
	if r.Status != gates.Fail {
		t.Fatalf("expected Fail, got %v", r.Status)
	}
	if !strings.Contains(strings.Join(r.Details, "\n"), "a -> b (top may not import leaf)") {
		t.Fatalf("violating edge string absent: %v", r.Details)
	}
}

// Scenario Outline: Malformed config — Fail with a config-error message
// distinct from a violation (no arrow).
func TestAccept_ImportGraph_MalformedConfigDistinctFromViolation(t *testing.T) {
	igFixture(t, "package a\n")
	bad := []config.Layer{{Name: "top", Paths: []string{"a/**"}, Allow: []string{"ghost"}}}
	r := igAccResult(t, igAccConfig(bad))
	if r.Status != gates.Fail {
		t.Fatalf("expected Fail, got %v", r.Status)
	}
	if !strings.HasPrefix(r.Message, "import_graph config:") || strings.Contains(r.Message, "->") {
		t.Fatalf("config error must be distinct from a violation edge: %q", r.Message)
	}
}
