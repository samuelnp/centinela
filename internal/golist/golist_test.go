package golist

import (
	"strings"
	"testing"
)

// chdirFixture writes a minimal two-package module and chdirs into it.
func chdirFixture(t *testing.T, goMod string) {
	t.Helper()
	d := chdirTemp(t)
	writeT(t, d, "go.mod", goMod)
	writeT(t, d, "b/b.go", "package b\n\nfunc B() {}\n")
	writeT(t, d, "a/a.go", "package a\n\nimport _ \"fixturemod/b\"\n")
}

func chdirBroken(t *testing.T) {
	t.Helper()
	writeT(t, chdirTemp(t), "go.mod", "not a go.mod\n")
}

func TestModulePath_Fixture(t *testing.T) {
	chdirFixture(t, "module fixturemod\n\ngo 1.21\n")
	m, err := ModulePath()
	if err != nil || m != "fixturemod" {
		t.Fatalf("ModulePath: got %q %v", m, err)
	}
}

func TestPackages_DecodesStreamedJSON(t *testing.T) {
	chdirFixture(t, "module fixturemod\n\ngo 1.21\n")
	pkgs, err := Packages()
	if err != nil {
		t.Fatal(err)
	}
	byPath := map[string]Pkg{}
	for _, p := range pkgs {
		byPath[p.ImportPath] = p
	}
	a, ok := byPath["fixturemod/a"]
	if !ok {
		t.Fatalf("package a not decoded: %#v", pkgs)
	}
	found := false
	for _, imp := range a.Imports {
		if imp == "fixturemod/b" {
			found = true
		}
	}
	if !found {
		t.Fatalf("a.Imports must include fixturemod/b: %#v", a.Imports)
	}
}

func TestPackages_ErrorSurfaced(t *testing.T) {
	chdirBroken(t)
	if _, err := Packages(); err == nil {
		t.Fatal("uncompilable module must surface a go list error, not empty success")
	}
}

func TestModulePath_ErrorSurfaced(t *testing.T) {
	chdirBroken(t)
	_, err := ModulePath()
	if err == nil || !strings.Contains(err.Error(), "go list -m") {
		t.Fatalf("module discovery error must be surfaced: %v", err)
	}
}

func TestFirstStderrLine(t *testing.T) {
	if got := firstStderrLine("\n  \nfirst real\nsecond\n"); got != "first real" {
		t.Fatalf("firstStderrLine: got %q", got)
	}
	if got := firstStderrLine("\n   \n\n"); got != "" {
		t.Fatalf("blank-only stderr must yield empty string, got %q", got)
	}
}
