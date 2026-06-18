package analyze

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// sampleInventory returns a small populated inventory for Save tests.
func sampleInventory() Inventory {
	return Inventory{
		SchemaVersion:   SchemaVersion,
		PrimaryLanguage: "Go",
		Languages:       []LanguageStat{{Name: "Go", FileCount: 2}},
		Manifests:       []Manifest{{Kind: "go-mod", Path: "go.mod", Build: "mod/path"}},
		Locales:         []string{"en", "es"},
		Packages:        []string{"a", "b"},
		Graph:           DependencyGraph{Kind: "go-packages", Module: "mod/path", Edges: []Edge{{From: "a", To: "b"}}},
	}
}

func TestSave_ByteStableAcrossReruns(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "sub", "analysis.json")
	inv := sampleInventory()
	if err := Save(p, inv); err != nil {
		t.Fatal(err)
	}
	first, _ := os.ReadFile(p)
	if err := Save(p, inv); err != nil {
		t.Fatal(err)
	}
	second, _ := os.ReadFile(p)
	if string(first) != string(second) {
		t.Fatalf("Save must be byte-identical across re-runs (AC-3)")
	}
	if !strings.HasSuffix(string(first), "}\n") {
		t.Fatalf("Save must end with a trailing newline, got %q", first[len(first)-3:])
	}
}

func TestSave_SchemaVersionPresent(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "analysis.json")
	if err := Save(p, sampleInventory()); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(p)
	if !strings.Contains(string(data), `"schemaVersion": 1`) {
		t.Fatalf("schemaVersion must be serialized (AC-8): %s", data)
	}
}

func TestSave_UnwritablePathErrors(t *testing.T) {
	// Parent path is a file, so MkdirAll/WriteFile fails: no partial artifact.
	dir := t.TempDir()
	blocker := filepath.Join(dir, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Save(filepath.Join(blocker, "analysis.json"), sampleInventory()); err == nil {
		t.Fatal("Save into a non-directory path must error")
	}
}
