package gates

import (
	"strings"
	"testing"
)

func TestStripModulePrefix(t *testing.T) {
	const mod = "github.com/samuelnp/centinela"
	if r, ok := stripModulePrefix(mod, mod); !ok || r != "" {
		t.Fatalf("module root: got %q,%v", r, ok)
	}
	if r, ok := stripModulePrefix(mod+"/internal/config", mod); !ok || r != "internal/config" {
		t.Fatalf("in-module: got %q,%v", r, ok)
	}
	if _, ok := stripModulePrefix(mod+"x/other", mod); ok {
		t.Fatal("segment-boundary substring must not match")
	}
	if _, ok := stripModulePrefix("fmt", mod); ok {
		t.Fatal("stdlib must be dropped")
	}
}

func TestScopePackages_FoldsTestImportsDropsExternal(t *testing.T) {
	const mod = "github.com/samuelnp/centinela"
	raw := []goListPkg{{
		ImportPath:   mod + "/internal/gates",
		Imports:      []string{"fmt", mod + "/internal/config"},
		TestImports:  []string{mod + "/internal/gates"}, // self -> dropped
		XTestImports: []string{mod + "/internal/workflow", "os"},
	}, {ImportPath: "golang.org/x/tools"}} // third-party pkg dropped entirely
	out := scopePackages(raw, mod)
	if len(out) != 1 || out[0].Path != "internal/gates" {
		t.Fatalf("expected 1 scoped pkg, got %+v", out)
	}
	got := strings.Join(out[0].Imports, ",")
	if !strings.Contains(got, "internal/config") || !strings.Contains(got, "internal/workflow") {
		t.Fatalf("test imports not folded: %q", got)
	}
	if strings.Contains(got, "fmt") || strings.Contains(got, "internal/gates") {
		t.Fatalf("stdlib/self should be dropped: %q", got)
	}
}
