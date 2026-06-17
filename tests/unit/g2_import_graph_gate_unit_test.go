package unit_test

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/gates"
)

// Acceptance spec: specs/g2-import-graph-gate.feature
//
// Tests-tier unit coverage exercises the gate's pure result semantics through
// the public gates package surface (RunWithFilter), independent of a real load.

func igConfig(layers []config.Layer) *config.Config {
	cfg := &config.Config{}
	cfg.Gates.ImportGraph = config.ImportGraphConfig{Enabled: true, Module: "m", Layers: layers}
	return cfg
}

func importGraphResult(cfg *config.Config) (gates.Result, bool) {
	for _, r := range gates.RunWithFilter(cfg, nil) {
		if r.Name == "import_graph" {
			return r, true
		}
	}
	return gates.Result{}, false
}

// Scenario: Gate explicitly disabled / no block present -> gate is omitted.
func TestImportGraph_DisabledOmitted(t *testing.T) {
	cfg := &config.Config{} // Enabled defaults to false
	if _, ok := importGraphResult(cfg); ok {
		t.Fatal("disabled gate must not appear in results")
	}
}

// Scenario: Block present with no layers defined -> Warn (not a silent Pass).
func TestImportGraph_EmptyMatrixWarns(t *testing.T) {
	r, ok := importGraphResult(igConfig(nil))
	if !ok || r.Status != gates.Warn {
		t.Fatalf("empty matrix should Warn, got ok=%v status=%v", ok, r.Status)
	}
	if !strings.Contains(strings.ToLower(r.Message), "empty") {
		t.Fatalf("warn should mention empty matrix, got %q", r.Message)
	}
}

// Scenario Outline: Malformed config -> Fail with "import_graph config:" and no arrow.
func TestImportGraph_MalformedConfigFails(t *testing.T) {
	bad := []config.Layer{{Name: "x", Paths: []string{"a/**"}, Allow: []string{"ghost"}}}
	r, ok := importGraphResult(igConfig(bad))
	if !ok || r.Status != gates.Fail {
		t.Fatalf("malformed config should Fail, got ok=%v status=%v", ok, r.Status)
	}
	if !strings.HasPrefix(r.Message, "import_graph config:") {
		t.Fatalf("config error prefix wrong: %q", r.Message)
	}
	if strings.Contains(r.Message, "->") {
		t.Fatalf("config error must not contain a violation arrow: %q", r.Message)
	}
}
