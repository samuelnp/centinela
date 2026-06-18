package integration_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/analyze"
)

// writeAnalyzeFixture builds a t.TempDir mini Go module with a package.json and a
// locales/ dir, returning the root. Source files double as read-only witnesses.
func writeAnalyzeFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if r, err := filepath.EvalSymlinks(root); err == nil {
		root = r
	}
	w := func(rel, body string) {
		p := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	w("go.mod", "module fixturemod\n\ngo 1.21\n")
	w("b/b.go", "package b\n\nfunc B() {}\n")
	w("a/a.go", "package a\n\nimport _ \"fixturemod/b\"\n")
	w("package.json", `{"scripts":{"build":"vite build","test":"vitest"},"dependencies":{"react":"18"}}`)
	w("locales/en.json", "{}")
	w("locales/es.json", "{}")
	return root
}

// analyzeIn chdirs into root (matching the binary's Analyze(".") semantics, so
// the golist-backed Go graph resolves against the fixture module, not the test's
// own CWD), runs Analyze + Save, and returns the written bytes.
func analyzeIn(t *testing.T, root string) []byte {
	t.Helper()
	orig, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(orig) })
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	inv, err := analyze.Analyze(".")
	if err != nil {
		t.Fatalf("analyze: %v", err)
	}
	if err := analyze.Save(analyze.DefaultOutPath, inv); err != nil {
		t.Fatalf("save: %v", err)
	}
	data, _ := os.ReadFile(analyze.DefaultOutPath)
	return data
}

func TestAnalyzeWritesCompleteInventory(t *testing.T) {
	root := writeAnalyzeFixture(t)
	var got analyze.Inventory
	data := analyzeIn(t, root)
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("inventory not valid JSON: %v", err)
	}
	if got.PrimaryLanguage != "Go" || got.SchemaVersion != analyze.SchemaVersion {
		t.Fatalf("head: %#v", got)
	}
	if got.Graph.Module != "fixturemod" || len(got.Graph.Edges) == 0 {
		t.Fatalf("Go graph must record module + edges: %#v", got.Graph)
	}
	if !hasNpm(got) {
		t.Fatalf("npm manifest scripts missing: %#v", got.Manifests)
	}
	if len(got.Locales) != 2 {
		t.Fatalf("locales: %v", got.Locales)
	}
}

func hasNpm(inv analyze.Inventory) bool {
	for _, m := range inv.Manifests {
		if m.Kind == "npm" && m.Build == "vite build" && m.Test == "vitest" {
			return true
		}
	}
	return false
}
