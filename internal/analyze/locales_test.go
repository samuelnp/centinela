package analyze

import (
	"os"
	"path/filepath"
	"testing"
)

func mkFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestDetectLocales_RootAndNestedDirs(t *testing.T) {
	root := t.TempDir()
	mkFile(t, filepath.Join(root, "locales", "en.json"), "{}")
	mkFile(t, filepath.Join(root, "locales", "es.json"), "{}")
	mkFile(t, filepath.Join(root, "src", "i18n", "pt-BR.json"), "{}")
	mkFile(t, filepath.Join(root, "locales", "README.md"), "x") // not a locale code
	got := detectLocales(root)
	want := map[string]bool{"en": true, "es": true, "pt-BR": true}
	if len(got) != len(want) {
		t.Fatalf("locales: got %v want %v", got, want)
	}
	for _, c := range got {
		if !want[c] {
			t.Fatalf("unexpected locale %q in %v", c, got)
		}
	}
	if got[0] != "en" || got[1] != "es" {
		t.Fatalf("locales must be sorted: %v", got)
	}
}

func TestDetectLocales_DirEntriesAsCodes(t *testing.T) {
	root := t.TempDir()
	mkFile(t, filepath.Join(root, "lang", "fr", "msg.po"), "x")
	got := detectLocales(root)
	if len(got) != 1 || got[0] != "fr" {
		t.Fatalf("locale subdir name must be detected: %v", got)
	}
}

func TestDetectLocales_NoI18nIsEmpty(t *testing.T) {
	root := t.TempDir()
	mkFile(t, filepath.Join(root, "main.go"), "package main")
	if got := detectLocales(root); len(got) != 0 {
		t.Fatalf("no-i18n repo must yield empty locales, got %v", got)
	}
}
