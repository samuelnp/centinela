package reconstruct

import (
	"testing"

	"github.com/samuelnp/centinela/internal/analyze"
)

func TestNewSignals_FlattensAndLowercases(t *testing.T) {
	in := analyze.Inventory{
		PrimaryLanguage: "Go",
		Packages:        []string{"Internal/Service"},
		Manifests: []analyze.Manifest{
			{Kind: "go-mod", Framework: "Cobra", Deps: []string{"GitHub.com/Spf13/Cobra"}},
			{Kind: "npm", Framework: "", Deps: []string{"Express"}},
		},
		Graph: analyze.DependencyGraph{Edges: []analyze.Edge{{From: "A", To: "Pkg/Calc"}}},
	}
	s := newSignals(in)
	if s.lang != "go" || s.pkgs[0] != "internal/service" {
		t.Fatalf("lang/pkg not lowercased: %+v", s)
	}
	if !s.hasDep("cobra") || !s.hasDep("express") {
		t.Fatalf("deps not flattened/lowercased: %v", s.deps)
	}
	if !s.hasFramework("cobra") || s.hasFramework("express") {
		t.Fatalf("framework filtering wrong (empty framework must be skipped): %v", s.frames)
	}
	if !s.hasIncoming("pkg/calc") || s.hasIncoming("missing") {
		t.Fatalf("incoming edge lookup wrong: %v", s.inEdges)
	}
	if len(s.kinds) != 2 {
		t.Fatalf("kinds not collected: %v", s.kinds)
	}
}

func TestAnyContainsAndContains(t *testing.T) {
	if !anyContains([]string{"a", "bcd"}, "cd") || anyContains(nil, "x") {
		t.Fatal("anyContains mismatch")
	}
	if !contains("hello", "ell") || contains("hello", "zz") {
		t.Fatal("contains mismatch")
	}
}
