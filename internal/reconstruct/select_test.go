package reconstruct

import (
	"testing"

	"github.com/samuelnp/centinela/internal/analyze"
)

// inv builds a minimal Inventory fixture for selection tests.
func inv(lang string, pkgs []string, edges ...analyze.Edge) analyze.Inventory {
	return analyze.Inventory{
		SchemaVersion: analyze.SchemaVersion, PrimaryLanguage: lang,
		Packages: pkgs, Graph: analyze.DependencyGraph{Edges: edges},
	}
}

func slugs(ts []Target) []string {
	out := make([]string, len(ts))
	for i, t := range ts {
		out[i] = t.Slug
	}
	return out
}

func TestSelect_GoNTierRolesAndSort(t *testing.T) {
	got := Select(inv("Go", []string{"internal/service", "cmd/app", "internal/handler"}))
	if want := []string{"cmd-app", "internal-handler", "internal-service"}; !equal(slugs(got), want) {
		t.Fatalf("slugs sorted wrong: %v", slugs(got))
	}
	byPkg := map[string]Role{}
	for _, x := range got {
		byPkg[x.Pkg] = x.Role
	}
	if byPkg["cmd/app"] != RoleCommand || byPkg["internal/handler"] != RoleEndpoint || byPkg["internal/service"] != RoleModule {
		t.Fatalf("role hints wrong: %+v", byPkg)
	}
}

func TestSelect_EmptyAndDocOnlyZeroTargets(t *testing.T) {
	if got := Select(inv("", nil)); len(got) != 0 {
		t.Fatalf("empty inventory must select 0 targets, got %d", len(got))
	}
	if got := Select(inv("Markdown", []string{"docs", "readme"})); len(got) != 0 {
		t.Fatalf("doc-only inventory must select 0 targets, got %v", slugs(got))
	}
}

func TestSelect_PolyglotEmptyGraphFromManifest(t *testing.T) {
	in := analyze.Inventory{SchemaVersion: analyze.SchemaVersion, PrimaryLanguage: "JavaScript",
		Packages:  []string{"src/api/users", "src/util"},
		Manifests: []analyze.Manifest{{Kind: "npm", Path: "package.json", Framework: "express", Deps: []string{"express"}}},
		Graph:     analyze.DependencyGraph{Edges: nil}}
	got := Select(in)
	if len(got) == 0 {
		t.Fatal("polyglot inventory with empty graph must still select targets")
	}
	found := false
	for _, x := range got {
		if x.Pkg == "src/api/users" && x.Role == RoleEndpoint {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected api endpoint target via manifest/package, got %+v", got)
	}
}

func TestSelect_GraphInEdgePromotesConsumedSurface(t *testing.T) {
	in := inv("Go", []string{"pkg/calc"}, analyze.Edge{From: "main", To: "pkg/calc"})
	got := Select(in)
	if len(got) != 1 || got[0].Role != RoleModule {
		t.Fatalf("consumed surface must be promoted as module, got %+v", got)
	}
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
