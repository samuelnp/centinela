package analyze

import (
	"path/filepath"
	"testing"
)

func TestAnalyze_PolyglotAssembly(t *testing.T) {
	root := t.TempDir()
	mkFile(t, filepath.Join(root, "main.go"), "package main")
	mkFile(t, filepath.Join(root, "app.js"), "console.log(1)")
	mkFile(t, filepath.Join(root, "lib.rb"), "puts 1")
	mkFile(t, filepath.Join(root, "package.json"), `{"scripts":{"test":"jest"},"dependencies":{"react":"18"}}`)
	mkFile(t, filepath.Join(root, "locales", "en.json"), "{}")
	inv, err := Analyze(root)
	if err != nil {
		t.Fatal(err)
	}
	if inv.SchemaVersion != SchemaVersion {
		t.Fatalf("schemaVersion: %d", inv.SchemaVersion)
	}
	names := map[string]bool{}
	for _, l := range inv.Languages {
		names[l.Name] = true
	}
	if !names["Go"] || !names["JavaScript"] || !names["Ruby"] {
		t.Fatalf("polyglot languages missing: %#v", inv.Languages)
	}
	if len(inv.Manifests) == 0 || len(inv.Locales) == 0 || len(inv.Packages) == 0 {
		t.Fatalf("sections must be populated: %#v", inv)
	}
}

func TestAnalyze_FailingSubDetectorStillValid(t *testing.T) {
	// A malformed package.json (a failing sub-detector) must not error the run;
	// the inventory is still valid and complete (AC-7).
	root := t.TempDir()
	mkFile(t, filepath.Join(root, "main.go"), "package main")
	mkFile(t, filepath.Join(root, "package.json"), "{ not json")
	inv, err := Analyze(root)
	if err != nil {
		t.Fatalf("malformed sub-detector must not error: %v", err)
	}
	if inv.PrimaryLanguage != "Go" || len(inv.Manifests) == 0 {
		t.Fatalf("inventory must still be populated: %#v", inv)
	}
}

func TestAnalyze_EmptyRepo(t *testing.T) {
	root := t.TempDir()
	mkFile(t, filepath.Join(root, "README.md"), "# docs")
	inv, err := Analyze(root)
	if err != nil {
		t.Fatal(err)
	}
	if inv.PrimaryLanguage != "" {
		t.Fatalf("empty repo primaryLanguage must be \"\", got %q", inv.PrimaryLanguage)
	}
	if len(inv.Manifests) != 0 || inv.Graph.Kind != "none" {
		t.Fatalf("empty repo manifests/graph must be empty: %#v", inv)
	}
}

func TestAnalyze_UnreadableRootErrors(t *testing.T) {
	if _, err := Analyze(filepath.Join(t.TempDir(), "nope")); err == nil {
		t.Fatal("unreadable root must be a hard error")
	}
}
