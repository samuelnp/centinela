package gates

import (
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

// igCfg builds a Config whose import-graph block uses the sample layers and an
// explicit module path so checkImportGraph never shells out for module
// discovery (those exec paths are covered by the load-tier fixture test).
func igCfg(module string, layers []config.Layer) *config.Config {
	cfg := &config.Config{}
	cfg.Gates.ImportGraph = config.ImportGraphConfig{Enabled: true, Module: module, Layers: layers}
	return cfg
}

func TestCheckImportGraph_EmptyMatrixWarns(t *testing.T) {
	r := checkImportGraph(igCfg("m", nil), nil)
	if r.Status != Warn || r.Name != "import_graph" {
		t.Fatalf("empty matrix should Warn, got %v %q", r.Status, r.Message)
	}
}

func TestCheckImportGraph_ConfigErrorFails(t *testing.T) {
	bad := []config.Layer{{Name: "x", Paths: []string{"a/**"}, Allow: []string{"ghost"}}}
	r := checkImportGraph(igCfg("m", bad), nil)
	if r.Status != Fail {
		t.Fatalf("expected Fail, got %v", r.Status)
	}
	if got := r.Message; got[:18] != "import_graph confi" {
		t.Fatalf("config error message prefix wrong: %q", got)
	}
}

func TestResolveModule(t *testing.T) {
	if m, err := resolveModule("  github.com/a/b  "); err != nil || m != "  github.com/a/b  " {
		t.Fatalf("explicit module returned verbatim: %q %v", m, err)
	}
	// Blank module triggers discovery via `go list -m`; inside this repo that
	// resolves to the centinela module path without error.
	m, err := resolveModule("")
	if err != nil || m == "" {
		t.Fatalf("blank module discovery failed: %q %v", m, err)
	}
}

func TestCheckImportGraph_PassOnRealModule(t *testing.T) {
	// Run against THIS module with a deliberately permissive single layer that
	// maps every package, so no edge can be forbidden -> Pass. Exercises the
	// full load -> scope -> check pipeline via real `go list -json`.
	layers := []config.Layer{{Name: "all", Paths: []string{"**"}, Allow: nil}}
	r := checkImportGraph(igCfg("", layers), nil)
	if r.Status != Pass {
		t.Fatalf("single all-layer should Pass, got %v: %q\n%v", r.Status, r.Message, r.Details)
	}
}

func TestCheckImportGraph_WarnOnUnmapped(t *testing.T) {
	// A layer matching nothing -> every package is unmapped -> Warn.
	layers := []config.Layer{{Name: "none", Paths: []string{"does/not/exist/**"}, Allow: nil}}
	r := checkImportGraph(igCfg("", layers), nil)
	if r.Status != Warn {
		t.Fatalf("expected Warn for all-unmapped, got %v: %q", r.Status, r.Message)
	}
	if len(r.Details) == 0 {
		t.Fatal("warn should list unmapped packages")
	}
}
