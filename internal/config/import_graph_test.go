package config

import (
	"testing"

	"github.com/BurntSushi/toml"
)

func TestNormalizeImportGraph_TrimsAndPreservesEmptyPaths(t *testing.T) {
	in := ImportGraphConfig{
		Module: "  github.com/x/y  ",
		Layers: []Layer{
			{Name: " config ", Paths: []string{" internal/config/** ", "  "}, Allow: []string{" ", "leaf"}},
			{Name: "empty", Paths: []string{"  ", ""}, Allow: nil},
		},
	}
	got := NormalizeImportGraph(in)
	if got.Module != "github.com/x/y" {
		t.Fatalf("module not trimmed: %q", got.Module)
	}
	if got.Layers[0].Name != "config" {
		t.Fatalf("layer name not trimmed: %q", got.Layers[0].Name)
	}
	if len(got.Layers[0].Paths) != 1 || got.Layers[0].Paths[0] != "internal/config/**" {
		t.Fatalf("paths not trimmed/filtered: %+v", got.Layers[0].Paths)
	}
	if len(got.Layers[0].Allow) != 1 || got.Layers[0].Allow[0] != "leaf" {
		t.Fatalf("allow not trimmed/filtered: %+v", got.Layers[0].Allow)
	}
	// Empty-path layer is PRESERVED (so the gate can report a config error).
	if len(got.Layers) != 2 || len(got.Layers[1].Paths) != 0 {
		t.Fatalf("empty-path layer should be preserved with zero paths: %+v", got.Layers)
	}
}

func TestImportGraph_TOMLDecodeRoundTrip(t *testing.T) {
	src := `
[gates.import_graph]
enabled = true
module = "github.com/samuelnp/centinela"

[[gates.import_graph.layers]]
name = "leaf"
paths = ["internal/config/**"]
allow = []

[[gates.import_graph.layers]]
name = "domain"
paths = ["internal/gates/**"]
allow = ["leaf"]
`
	var cfg Config
	if _, err := toml.Decode(src, &cfg); err != nil {
		t.Fatalf("decode: %v", err)
	}
	g := cfg.Gates.ImportGraph
	if !g.Enabled || g.Module != "github.com/samuelnp/centinela" {
		t.Fatalf("scalar decode wrong: %+v", g)
	}
	if len(g.Layers) != 2 {
		t.Fatalf("expected 2 layers, got %d", len(g.Layers))
	}
	if g.Layers[1].Name != "domain" || len(g.Layers[1].Allow) != 1 || g.Layers[1].Allow[0] != "leaf" {
		t.Fatalf("layer decode wrong: %+v", g.Layers[1])
	}
}

func TestApplyDefaults_NormalizesImportGraphModule(t *testing.T) {
	cfg := &Config{}
	cfg.Gates.ImportGraph.Module = "  github.com/a/b  "
	applyDefaults(cfg)
	if cfg.Gates.ImportGraph.Module != "github.com/a/b" {
		t.Fatalf("applyDefaults did not normalize module: %q", cfg.Gates.ImportGraph.Module)
	}
}
